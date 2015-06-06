package main

import (
	"html/template"
	"net/http"

	"github.com/BenLubar/espeak"
	"github.com/BenLubar/espeak/wav"
)

func main() {
	http.HandleFunc("/", handler)

	if err := http.ListenAndServe(":6060", nil); err != nil {
		panic(err)
	}
}

var tmpl = template.Must(template.New("tts").Parse(`<!DOCTYPE html>
<html>
<head>
<title>Text to Speech</title>
</head>
<body>
<form action="/" method="get">
<textarea id="text" name="text">{{.Req.FormValue "text"}}</textarea><br>
<label for="voice">Voice:</label>
<select id="voice" name="voice">
{{with $base := .}}
{{range .Voices}}
<option value="{{.Name}}"{{if eq .Name ($base.Req.FormValue "voice")}} selected{{end}}>{{.Name}}{{range .Languages}} ({{.Language}}){{end}}</option>
{{end}}
{{end}}
</select><br>
<input type="hidden" name="tts" value="1">
<input type="submit" value="Text to Speech">
</form>
</body>
</html>
`))

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.FormValue("tts") == "" {
		if err := tmpl.Execute(w, &struct {
			Req    *http.Request
			Voices []*espeak.Voice
		}{r, wav.AllVoices()}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	voices := wav.FindVoices(r.FormValue("voice"), "", espeak.NoGender, 0)
	var voice *espeak.Voice
	if len(voices) != 0 {
		voice = voices[0]
	}

	w.Header().Set("Content-Type", "audio/wav")
	if err := wav.Synth(w, voice, r.FormValue("text")); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
