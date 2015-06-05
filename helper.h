#include <espeak/speak_lib.h>

espeak_EVENT_TYPE eventType(const espeak_EVENT *);
int eventNumber(const espeak_EVENT *);
const char *eventName(const espeak_EVENT *);
const char *eventString(const espeak_EVENT *);
void uriSetCallback();
