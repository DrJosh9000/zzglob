// The zzdot command parses a glob pattern, and prints the underlying state
// machine in GraphViz DOT language.
//
// Example:
//
//	$ zzdot '**/*_test.go'
//	digraph {
//		rankdir=LR;
//		comment="input pattern: \"**/*_test.go\" parser config: {allowEscaping:true allowQuestion:true allowStar:true allowDoubleStar:true allowAlternation:true allowCharClass:true swapSlashes:false expandTilde:true}";
//		initial [label="", style=invis];
//		initial -> state_0x140000be020 [label=""];
//		state_0x140000be020 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be020 -> state_0x140000be040 [label="<nil>"];
//		state_0x140000be020 -> state_0x140000be080 [label="<nil>"];
//		state_0x140000be040 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be040 -> state_0x140000be040 [label="*"];
//		state_0x140000be040 -> state_0x140000be0c0 [label="_"];
//		state_0x140000be080 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be080 -> state_0x140000be080 [label="**"];
//		state_0x140000be080 -> state_0x140000be040 [label="/"];
//		state_0x140000be0c0 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be0c0 -> state_0x140000be0e0 [label="t"];
//		state_0x140000be0e0 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be0e0 -> state_0x140000be100 [label="e"];
//		state_0x140000be100 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be100 -> state_0x140000be120 [label="s"];
//		state_0x140000be120 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be120 -> state_0x140000be140 [label="t"];
//		state_0x140000be140 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be140 -> state_0x140000be160 [label="."];
//		state_0x140000be160 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be160 -> state_0x140000be180 [label="g"];
//		state_0x140000be180 [label="", shape=circle, style=filled, fillcolor=white];
//		state_0x140000be180 -> state_0x140000be1a0 [label="o"];
//		state_0x140000be1a0 [label="", shape=doublecircle, style=filled, fillcolor=white];
//	}
package main

import (
	"fmt"
	"os"

	"drjosh.dev/zzglob"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s pattern\n", os.Args[0])
		os.Exit(1)
	}

	p, err := zzglob.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse pattern %q: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	if err := p.WriteDot(os.Stdout, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't write Dot output: %v\n", err)
		os.Exit(1)
	}
}
