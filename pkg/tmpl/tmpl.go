package tmpl

type templateContext struct {
	Query engine.Query
	Results  []mixer.Result
	Form     formValues
}


