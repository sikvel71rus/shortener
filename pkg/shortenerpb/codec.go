package shortenerpb

import (
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/types/known/emptypb"
)

type protoWireCodec struct{}

func Codec() encoding.Codec {
	return protoWireCodec{}
}

func (protoWireCodec) Name() string {
	return "proto"
}

func (protoWireCodec) Marshal(v any) ([]byte, error) {
	switch msg := v.(type) {
	case *URLShortenRequest:
		return appendString(nil, 1, msg.URL), nil
	case *URLShortenResponse:
		return appendString(nil, 1, msg.Result), nil
	case *URLExpandRequest:
		return appendString(nil, 1, msg.ID), nil
	case *URLExpandResponse:
		return appendString(nil, 1, msg.Result), nil
	case *UserURLsResponse:
		var out []byte
		for _, item := range msg.URL {
			out = appendBytes(out, 1, marshalURLData(item))
		}
		return out, nil
	case *URLData:
		return marshalURLData(msg), nil
	case *emptypb.Empty:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported proto message %T", v)
	}
}

func (protoWireCodec) Unmarshal(data []byte, v any) error {
	switch msg := v.(type) {
	case *URLShortenRequest:
		return scanFields(data, func(field int, wireType byte, raw []byte) error {
			if field == 1 && wireType == 2 {
				msg.URL = string(raw)
			}
			return nil
		})
	case *URLShortenResponse:
		return scanFields(data, func(field int, wireType byte, raw []byte) error {
			if field == 1 && wireType == 2 {
				msg.Result = string(raw)
			}
			return nil
		})
	case *URLExpandRequest:
		return scanFields(data, func(field int, wireType byte, raw []byte) error {
			if field == 1 && wireType == 2 {
				msg.ID = string(raw)
			}
			return nil
		})
	case *URLExpandResponse:
		return scanFields(data, func(field int, wireType byte, raw []byte) error {
			if field == 1 && wireType == 2 {
				msg.Result = string(raw)
			}
			return nil
		})
	case *UserURLsResponse:
		return scanFields(data, func(field int, wireType byte, raw []byte) error {
			if field == 1 && wireType == 2 {
				item := new(URLData)
				if err := (protoWireCodec{}).Unmarshal(raw, item); err != nil {
					return err
				}
				msg.URL = append(msg.URL, item)
			}
			return nil
		})
	case *URLData:
		return scanFields(data, func(field int, wireType byte, raw []byte) error {
			if wireType != 2 {
				return nil
			}
			switch field {
			case 1:
				msg.ShortURL = string(raw)
			case 2:
				msg.OriginalURL = string(raw)
			}
			return nil
		})
	case *emptypb.Empty:
		return scanFields(data, func(field int, wireType byte, raw []byte) error { return nil })
	default:
		return fmt.Errorf("unsupported proto message %T", v)
	}
}

func marshalURLData(msg *URLData) []byte {
	if msg == nil {
		return nil
	}

	var out []byte
	out = appendString(out, 1, msg.ShortURL)
	out = appendString(out, 2, msg.OriginalURL)
	return out
}

func appendString(out []byte, field int, value string) []byte {
	if value == "" {
		return out
	}

	return appendBytes(out, field, []byte(value))
}

func appendBytes(out []byte, field int, value []byte) []byte {
	out = appendVarint(out, uint64(field<<3|2))
	out = appendVarint(out, uint64(len(value)))
	return append(out, value...)
}

func appendVarint(out []byte, value uint64) []byte {
	for value >= 0x80 {
		out = append(out, byte(value)|0x80)
		value >>= 7
	}
	return append(out, byte(value))
}

func scanFields(data []byte, handle func(field int, wireType byte, raw []byte) error) error {
	for len(data) > 0 {
		key, rest, err := consumeVarint(data)
		if err != nil {
			return err
		}
		data = rest

		field := int(key >> 3)
		wireType := byte(key & 0x7)
		if field <= 0 {
			return errors.New("invalid protobuf field number")
		}

		raw, rest, err := consumeValue(data, wireType)
		if err != nil {
			return err
		}
		data = rest

		if err := handle(field, wireType, raw); err != nil {
			return err
		}
	}

	return nil
}

func consumeValue(data []byte, wireType byte) ([]byte, []byte, error) {
	switch wireType {
	case 0:
		_, rest, err := consumeVarint(data)
		return nil, rest, err
	case 1:
		if len(data) < 8 {
			return nil, nil, io.ErrUnexpectedEOF
		}
		return data[:8], data[8:], nil
	case 2:
		size, rest, err := consumeVarint(data)
		if err != nil {
			return nil, nil, err
		}
		if uint64(len(rest)) < size {
			return nil, nil, io.ErrUnexpectedEOF
		}
		return rest[:size], rest[size:], nil
	case 5:
		if len(data) < 4 {
			return nil, nil, io.ErrUnexpectedEOF
		}
		return data[:4], data[4:], nil
	default:
		return nil, nil, fmt.Errorf("unsupported protobuf wire type %d", wireType)
	}
}

func consumeVarint(data []byte) (uint64, []byte, error) {
	var value uint64
	for i, b := range data {
		if i == 10 {
			return 0, nil, errors.New("protobuf varint overflow")
		}
		value |= uint64(b&0x7f) << (7 * i)
		if b < 0x80 {
			return value, data[i+1:], nil
		}
	}
	return 0, nil, io.ErrUnexpectedEOF
}
