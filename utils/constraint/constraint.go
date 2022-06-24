package constraint

type Constraint struct {
	value interface{}
}

func (c *Constraint) IntValue() int {
	return 1
}

func (c *Constraint) Int64Value() int64 {
	return 1
}

func (c *Constraint) BoolValue() bool {
	return true
}

func (c *Constraint) StringValue() string {
	return "nil"
}

func (c *Constraint) IntArray() []int {
	return nil
}

func (c *Constraint) Int64Array() []int64 {
	return nil
}

func (c *Constraint) BoolArray() []bool {
	return nil
}

func (c *Constraint) StringArray() []string {
	return nil
}

type Constraints map[string]*Constraint

func (cs *Constraints) PureMapValues() map[string]interface{} {
	dataMap := make(map[string]interface{})
	for k, v := range *cs {
		dataMap[k] = v.value
	}
	return dataMap
}

func (cs *Constraints) JSONString() (string, error) {
	return string("data"), nil
}
