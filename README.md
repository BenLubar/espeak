# espeak

[![GoDoc](https://godoc.org/gopkg.in/BenLubar/espeak.v2?status.svg)](https://godoc.org/gopkg.in/BenLubar/espeak.v2) [![Maintainability](https://api.codeclimate.com/v1/badges/31ef58e637d3b5d6576c/maintainability)](https://codeclimate.com/github/BenLubar/espeak/maintainability) [![Go Report Card](https://goreportcard.com/badge/gopkg.in/BenLubar/espeak.v2)](https://goreportcard.com/report/gopkg.in/BenLubar/espeak.v2)

Package espeak is a wrapper around espeak-ng that works both natively and in gopherjs with the same API. espeak-ng is an open source text to speech library that has over one hundred voices and languages and supports speech synthesis markup language (SSML).

## To download this package:

```
go get -u gopkg.in/BenLubar/espeak.v2
```

## Looking for an older version?

The original implementation of this package from 2015 is still available at [`gopkg.in/BenLubar/espeak.v1`](https://gopkg.in/BenLubar/espeak.v1).

## Special thanks

- [espeak-ng](https://github.com/espeak-ng/espeak-ng) (text to speech)
- [emscripten](https://github.com/kripken/emscripten) (C to JavaScript)
- [gopherjs](https://github.com/gopherjs/gopherjs) (Go to JavaScript)

## Want to repurpose my code?

You may reuse any code in this repository for any purpose, with the exception of `libespeak-ng.inc.js`, which is a compiled version of GPLv3-licensed code from espeak-ng.

Compiled versions of this package use GPLv3 code and therefore must be used under a GPLv3-compatible license.
