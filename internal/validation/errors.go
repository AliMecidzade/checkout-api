package validation

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

func (e FieldError) Error() string {
	return e.Field + ": " + e.Message

}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func (e *ValidationError) With(field, rule, msg string) *ValidationError {
	e.Fields = append(e.Fields, FieldError{Field: field, Rule: rule, Message: msg})
	return e
}
