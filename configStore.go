package main

type configStorage map[string]interface{}

func (c configStorage) Str(name string) string {
	v, ok := c[name]

	if !ok {
		return ""
	}

	if sv, ok := v.(string); ok {
		return sv
	}

	return ""
}

func (c configStorage) Int64(name string) int64 {
	v, ok := c[name]

	if !ok {
		return 0
	}

	if sv, ok := v.(int64); ok {
		return sv
	}

	return 0
}
