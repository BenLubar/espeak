Package espeak provides bindings for eSpeak <http://espeak.sf.net>.

Please note that functions in this package should only be called from one
goroutine at a time as libespeak uses static variables for most of its
returned values.

---

Although `github.com/BenLubar/espeak` is licensed under the MIT license,
`libespeak` is licensed under the GPL. This means that the compiled version
of this package can only be used by GPL-compatible programs. Sorry, but there's
nothing I can do about it.
