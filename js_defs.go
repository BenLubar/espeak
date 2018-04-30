// +build ignore

package espeak

/*
#include <espeak-ng/espeak_ng.h>

typedef int assert_pointer_is_32_bits[(sizeof(void*)-4)*(4-sizeof(void*))];

#define sizeof_espeak_EVENT sizeof(espeak_EVENT)
#define offsetof_espeak_EVENT_type offsetof(espeak_EVENT, type)
#define offsetof_espeak_EVENT_text_position offsetof(espeak_EVENT, text_position)
#define offsetof_espeak_EVENT_length offsetof(espeak_EVENT, length)
#define offsetof_espeak_EVENT_audio_position offsetof(espeak_EVENT, audio_position)
#define offsetof_espeak_EVENT_number offsetof(espeak_EVENT, id.number)
#define offsetof_espeak_EVENT_name offsetof(espeak_EVENT, id.name)
#define offsetof_espeak_EVENT_string offsetof(espeak_EVENT, id.string)

#define sizeof_espeak_VOICE sizeof(espeak_VOICE)
#define offsetof_espeak_VOICE_name offsetof(espeak_VOICE, name)
#define offsetof_espeak_VOICE_languages offsetof(espeak_VOICE, languages)
#define offsetof_espeak_VOICE_identifier offsetof(espeak_VOICE, identifier)
#define offsetof_espeak_VOICE_gender offsetof(espeak_VOICE, gender)
#define offsetof_espeak_VOICE_age offsetof(espeak_VOICE, age)
#define offsetof_espeak_VOICE_variant offsetof(espeak_VOICE, variant)
*/
import "C"

const outputModeSynchronous = C.ENOUTPUT_MODE_SYNCHRONOUS
const sOK = C.ENS_OK
const posCharacter = C.POS_CHARACTER

const espeakCHARS_UTF8 = C.espeakCHARS_UTF8
const espeakSSML = C.espeakSSML

const espeakRATE = C.espeakRATE
const espeakVOLUME = C.espeakVOLUME
const espeakPITCH = C.espeakPITCH
const espeakRANGE = C.espeakRANGE

const espeakEVENT_LIST_TERMINATED = C.espeakEVENT_LIST_TERMINATED
const espeakEVENT_WORD = C.espeakEVENT_WORD
const espeakEVENT_SENTENCE = C.espeakEVENT_SENTENCE
const espeakEVENT_MARK = C.espeakEVENT_MARK
const espeakEVENT_PLAY = C.espeakEVENT_PLAY
const espeakEVENT_END = C.espeakEVENT_END
const espeakEVENT_MSG_TERMINATED = C.espeakEVENT_MSG_TERMINATED
const espeakEVENT_PHONEME = C.espeakEVENT_PHONEME
const espeakEVENT_SAMPLERATE = C.espeakEVENT_SAMPLERATE

const eventSize = C.sizeof_espeak_EVENT
const eventTypeOffset = C.offsetof_espeak_EVENT_type
const eventTextPositionOffset = C.offsetof_espeak_EVENT_text_position
const eventLengthOffset = C.offsetof_espeak_EVENT_length
const eventAudioPositionOffset = C.offsetof_espeak_EVENT_audio_position
const eventNumberOffset = C.offsetof_espeak_EVENT_number
const eventNameOffset = C.offsetof_espeak_EVENT_name
const eventStringOffset = C.offsetof_espeak_EVENT_string

const voiceSize = C.sizeof_espeak_VOICE
const voiceNameOffset = C.offsetof_espeak_VOICE_name
const voiceLanguagesOffset = C.offsetof_espeak_VOICE_languages
const voiceIdentifierOffset = C.offsetof_espeak_VOICE_identifier
const voiceGenderOffset = C.offsetof_espeak_VOICE_gender
const voiceAgeOffset = C.offsetof_espeak_VOICE_age
const voiceVariantOffset = C.offsetof_espeak_VOICE_variant
