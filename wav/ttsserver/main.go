package main

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/BenLubar/espeak"
	"github.com/BenLubar/espeak/wav"
)

func main() {
	http.HandleFunc("/", handler)

	if err := http.ListenAndServe(":8050", nil); err != nil {
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
<option value="{{.Name}}"{{if eq .Name (or ($base.Req.FormValue "voice") "default")}} selected{{end}}>{{.Name}}{{range .Languages}} ({{.Language}}){{end}}</option>
{{end}}
{{end}}
</select><br>
<label for="pitch">Pitch:</label>
<input type="range" id="pitch" name="pitch" value="{{or (.Req.FormValue "pitch") "50"}}" min="0" max="100"><br>
<label for="range">Range:</label>
<input type="range" id="range" name="range" value="{{or (.Req.FormValue "range") "50"}}" min="0" max="100"><br>
<label for="rate">Rate:</label>
<input type="range" id="rate" name="rate" value="{{or (.Req.FormValue "rate") "175"}}" min="80" max="450"><br>
<input type="submit" value="Text to Speech">
</form>
{{if .Req.FormValue "text"}}<audio src="/?voice={{or (.Req.FormValue "voice") "default"}}&amp;text={{.Req.FormValue "text"}}&amp;pitch={{or (.Req.FormValue "pitch") 50}}&amp;range={{or (.Req.FormValue "range") 50}}&amp;rate={{or (.Req.FormValue "rate") 175}}&amp;tts=1" type="audio/wav" autoplay controls></audio>{{end}}
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

	pitch, err := strconv.Atoi(r.FormValue("pitch"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pitchRange, err := strconv.Atoi(r.FormValue("range"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rate, err := strconv.Atoi(r.FormValue("rate"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "audio/wav")
	if err := wav.Synth(w, voice, r.FormValue("text"), pitch, pitchRange, rate); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
