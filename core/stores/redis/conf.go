package redis

import "errors"

var (
	ErrEmptyHost = errors.New("empty redis host")
	ErrEmptyType = errors.New("empty redis type")
	ErrEmptyKey  = errors.New("empty redis key")
)

type (
	Conf struct {
		Host string
		Type string `json:",default=node,options=node|cluster"`
		Pass string `json:",optional"`
	}

	KeyConf struct {
		Conf
		Key string `json:",optional"`
	}
)

func (rc Conf) NewRedis() (Node, error) {
	return NewRedis(rc.Host, rc.Type, rc.Pass)
}

func (rc Conf) Validate() error {
	if len(rc.Host) == 0 {
		return ErrEmptyHost
	}

	if len(rc.Type) == 0 {
		return ErrEmptyType
	}

	return nil
}

func (rkc KeyConf) Validate() error {
	if err := rkc.Conf.Validate(); err != nil {
		return err
	}

	if len(rkc.Key) == 0 {
		return ErrEmptyKey
	}

	return nil
}
