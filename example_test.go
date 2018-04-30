package espeak_test

import (
	"os"

	"gopkg.in/BenLubar/espeak.v2"
)

func Example_ssml() {
	const ssml = `<?xml version="1.0"?>
<speak version="1.1"
	xmlns="http://www.w3.org/2001/10/synthesis"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://www.w3.org/2001/10/synthesis
		http://www.w3.org/TR/speech-synthesis11/synthesis.xsd"
	xml:lang="en-US">
	<!--
	Dialogue from Confessor's Stronghold, part of Guild Wars 2 Living World Season 3 Episode 1.
	Used for demonstration purposes only. Copyright 2016 ArenaNet LLC.
	-->
	<voice gender="male" languages="en:en-GB">
		<s>Nice work<sub alias="">,</sub> Commander. Now if I could persuade you to take care of the others...</s>
	</voice>
	<voice gender="female"><prosody pitch="-30%">
		<s>Interesting. <break/> Countering the magic of these bloodstones returns whatever magical properties it absorbed.</s>
		<s>I know a certain <prosody rate="85%">big-eared asura</prosody> who'd <emphasis level="strong">love</emphasis> to be here to study this...</s>
	</prosody></voice>
	<voice gender="female" variant="2">
		<s>I <prosody pitch="+65%">heard</prosody> <prosody pitch="+50%" range="-90%" rate="+50%">that!</prosody></s>
	</voice>
</speak>`

	var ctx espeak.Context
	ctx.SynthesizeText(ssml)

	f, _ := os.Create("example-ssml.wav")
	defer f.Close()
	ctx.WriteTo(f)

	// Output:
}
