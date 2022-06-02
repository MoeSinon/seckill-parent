package trigger

type trigger struct {
	Name  string
	Value interface{}
}

type com interface {
	Set(name string, value interface{}) error
	Del(name string) error
	Get(name string) (interface{}, error)
}

func Newtrigger() *trigger {
	return &trigger{}
}

func (t *trigger) Set(name string, value interface{}) {
	t.Name = name
	t.Value = value
}

func (t *trigger) Get(name string) interface{} {
	return t.Value
}

func (t *trigger) Del(name string) {
	t.Name = ""
	t.Value = nil
}
