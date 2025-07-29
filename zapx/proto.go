package zapx

import (
	"unsafe"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type protoStringer struct {
	inner   proto.Message
	options ProtoFieldOptions
}

func (s protoStringer) String() string {
	if s.inner == nil {
		return "<nil>"
	}

	if s.options.MarshalerKind == ProtoFieldJSONMarshaler {
		data, err := s.options.ToProtoJSONMarshalOptions().Marshal(s.inner)
		if err != nil {
			return "<error marshaling proto to JSON: " + err.Error() + ">"
		}
		return string(data)
	}

	data, err := s.options.ToProtoTextMarshalOptions().Marshal(s.inner)
	if err != nil {
		return "<error marshaling proto to text: " + err.Error() + ">"
	}

	return unsafe.String(unsafe.SliceData(data), len(data))
}

type ProtoFieldMarshalerKind string

const (
	ProtoFieldTextMarshaler ProtoFieldMarshalerKind = "text"
	ProtoFieldJSONMarshaler ProtoFieldMarshalerKind = "json"
)

type ProtoFieldOptions struct {
	// Indent is the string used for indentation in pretty printing,
	// if empty, defaults is no indentation (compact format).
	Indent string

	// Resolver is used to resolve message types and extensions.
	Resolver interface {
		protoregistry.ExtensionTypeResolver
		protoregistry.MessageTypeResolver
	}

	// MarshalerKind specifies the kind of marshaling to use for the proto field,
	// either zapx.ProtoFieldTextMarshaler or zapx.ProtoFieldJSONMarshaler, defaults
	// is text marshaling.
	MarshalerKind ProtoFieldMarshalerKind

	// AllowPartial allows messages that have missing required fields to marshal
	// without returning an error. If AllowPartial is false,
	// Marshal will return error if there are any missing required fields. Defaults
	// is true.
	AllowPartial bool

	// EmitUnpopulated specifies whether unpopulated fields should be emitted,
	// defaults to false (applies to JSON marshaling only).
	EmitUnpopulated bool
}

func (opts *ProtoFieldOptions) ToProtoJSONMarshalOptions() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Resolver:        opts.Resolver,
		Indent:          opts.Indent,
		AllowPartial:    opts.AllowPartial,
		EmitUnpopulated: opts.EmitUnpopulated,
	}
}

func (opts *ProtoFieldOptions) ToProtoTextMarshalOptions() prototext.MarshalOptions {
	multiline := false
	if opts.Indent != "" {
		multiline = true
	}

	return prototext.MarshalOptions{
		Resolver:     opts.Resolver,
		Indent:       opts.Indent,
		Multiline:    multiline,
		AllowPartial: opts.AllowPartial,
	}
}

type ProtoFieldOption interface {
	Apply(*ProtoFieldOptions)
}

type ProtoFieldOptionFunc func(*ProtoFieldOptions)

func (f ProtoFieldOptionFunc) Apply(opts *ProtoFieldOptions) {
	f(opts)
}

// ProtoPretty can be used to render fields on protobuf messages in a pretty format
// with indentation and multiline support.
//
//	zapx.Proto("field", myProtoMessage, zapx.ProtoPretty())
func ProtoPretty() ProtoFieldOption {
	return ProtoFieldOptionFunc(func(opts *ProtoFieldOptions) {
		opts.Indent = "  "
	})
}

// ProtoMarshalerJSON can be used to render fields on protobuf messages in a pretty format
// with indentation and multiline support.
//
//	zapx.Proto("field", myProtoMessage, zapx.ProtoMarshalerJSON())
func ProtoMarshalerJSON() ProtoFieldOption {
	return ProtoFieldOptionFunc(func(opts *ProtoFieldOptions) {
		opts.MarshalerKind = ProtoFieldJSONMarshaler
	})
}

// ProtoResolver can be used to set a custom resolver for the proto field.
//
//	zapx.Proto("field", myProtoMessage, zapx.ProtoResolver(myResolver))
func ProtoResolver(resolver interface {
	protoregistry.ExtensionTypeResolver
	protoregistry.MessageTypeResolver
}) ProtoFieldOption {
	return ProtoFieldOptionFunc(func(opts *ProtoFieldOptions) {
		opts.Resolver = resolver
	})
}

// Proto can be used to create a zap.Field that marshals a protobuf message
// into a string representation. The field can be rendered in either JSON or text format,
// depending on the options provided.
//
// This is a finer control version of zap.Reflect when you deal with protobuf messages. It
// is lazy loaded deferring the marshaling until the field is actually needed for a logger.
//
// By default shows the in compact text format, but you can use the ProtoPretty option to
// render it in a pretty format with indentation and multiline support:
//
//	zapx.Proto("field", myProtoMessage, zapx.ProtoPretty())
//
// You set it to render in JSON format by using the [ProtoMarshalerJSON] option:
//
//	zapx.Proto("field", myProtoMessage, zapx.ProtoMarshalerJSON())
//
// Note: Think about using [ProtoJSON] instead of this function if you want to render
// the protobuf message in JSON format.
//
// Finally you can use the [ProtoResolver] option to set a custom resolver for the proto field,
// which is useful if you have custom message types or extensions that need to be resolved
// during marshaling:
//
//	zapx.Proto("field", myProtoMessage, zapx.ProtoResolver(myResolver))
func Proto(name string, value proto.Message, opts ...ProtoFieldOption) zap.Field {
	options := ProtoFieldOptions{
		MarshalerKind: ProtoFieldTextMarshaler,
		AllowPartial:  true,
	}
	for _, opt := range opts {
		opt.Apply(&options)
	}

	return zap.Stringer(name, protoStringer{inner: value, options: options})
}

// ProtoJSON is a convenience function that creates a zap.Field for a protobuf message
// in JSON format, convenience for `zapx.Proto(name, value, zapx.ProtoMarshalerJSON())`.
//
// See [Proto] for full documentation on how to use this function.
func ProtoJSON(name string, value proto.Message, opts ...ProtoFieldOption) zap.Field {
	return Proto(name, value, append(opts, ProtoMarshalerJSON())...)
}
