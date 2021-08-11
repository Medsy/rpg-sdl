package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"rpg-sdl/game"
	"rpg-sdl/ui2d"
	"runtime"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		fmt.Println("starting CPU profiling...")
		defer pprof.StopCPUProfile()
	}
	runtime.LockOSThread()
	g := game.NewGame(1, "game/maps/level1.map")

	go func() {
		g.Run()
	}()
	ui := ui2d.NewUI(g.InputChan, g.LevelChans[0])
	ui.GetInput()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date stats
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		fmt.Println("starting memory profiling...")
	}
}
