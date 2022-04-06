package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stringx"
)

func TestRedisConf(t *testing.T) {
	tests := []struct {
		name string
		Conf
		ok bool
	}{
		{
			name: "missing host",
			Conf: Conf{
				Host: "",
				Type: NodeType,
				Pass: "",
			},
			ok: false,
		},
		{
			name: "missing type",
			Conf: Conf{
				Host: "localhost:6379",
				Type: "",
				Pass: "",
			},
			ok: false,
		},
		{
			name: "ok",
			Conf: Conf{
				Host: "localhost:6379",
				Type: NodeType,
				Pass: "",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(stringx.RandId(), func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.Conf.Validate())
				client, err := test.Conf.NewRedis()
				assert.NoError(t, err)
				assert.NotNil(t, client)
			} else {
				assert.NotNil(t, test.Conf.Validate())
			}
		})
	}
}

func TestRedisKeyConf(t *testing.T) {
	tests := []struct {
		name string
		KeyConf
		ok bool
	}{
		{
			name: "missing host",
			KeyConf: KeyConf{
				Conf: Conf{
					Host: "",
					Type: NodeType,
					Pass: "",
				},
				Key: "foo",
			},
			ok: false,
		},
		{
			name: "missing key",
			KeyConf: KeyConf{
				Conf: Conf{
					Host: "localhost:6379",
					Type: NodeType,
					Pass: "",
				},
				Key: "",
			},
			ok: false,
		},
		{
			name: "ok",
			KeyConf: KeyConf{
				Conf: Conf{
					Host: "localhost:6379",
					Type: NodeType,
					Pass: "",
				},
				Key: "foo",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.KeyConf.Validate())
			} else {
				assert.NotNil(t, test.KeyConf.Validate())
			}
		})
	}
}
