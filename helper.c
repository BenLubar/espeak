#include <string.h>
#include "helper.h"
#include "_cgo_export.h"

// In Go, type is a keyword, so we need this helper function.
espeak_EVENT_TYPE eventType(const espeak_EVENT *event) { return event->type; }

// Go doesn't have unions, so we do this part in C.
int eventNumber(const espeak_EVENT *event) { return event->id.number; }
const char *eventName(const espeak_EVENT *event) { return event->id.name; }
const char *eventString(const espeak_EVENT *event)
{
	// The string can use all 8 bytes of the buffer. Pointer arithmetic
	// is a lot harder in Go, so we convert it here.
	static char string[9];

	memcpy(string, event->id.string, 8);
	string[8] = 0;

	return string;
}

// We can't do const arguments in Go, so we need this helper function.
int uriCallbackHelper(int type, const char *uri, const char *base)
{
	return uriCallback((char *) uri, (char *) base);
}

// The argument to espeak_SetUriCallback is an anonymous type, so we can't cast it in Go.
void uriSetCallback() { espeak_SetUriCallback(uriCallbackHelper); }
