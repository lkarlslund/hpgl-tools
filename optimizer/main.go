package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/fzipp/geom"
	"github.com/spf13/pflag"
)

type Lines struct {
	Vectors []geom.Vec2
}

func (l Lines) Reverse() Lines {
	var newlines Lines
	newlines.Vectors = make([]geom.Vec2, 0, len(l.Vectors))
	for i := len(l.Vectors) - 1; i >= 0; i-- {
		newlines.Vectors = append(newlines.Vectors, l.Vectors[i])
	}
	return newlines
}

type Plot struct {
	lines []Lines
}

func (p Plot) MoveDistances(currentpos geom.Vec2) float32 {
	var movedist float32
	for _, line := range p.lines {
		movedist += line.Vectors[0].SqDist(currentpos)
		currentpos = line.Vectors[len(line.Vectors)-1]
	}
	return movedist
}

func (p Plot) Optimize(optimizedpos *geom.Vec2, reversible bool) Plot {
	var neworder []Lines

	used := make([]bool, len(p.lines))

	for iteration := 0; iteration < len(p.lines); iteration++ {
		bestindex := -1
		bestdistance := float32(-1)
		bestisreversed := false

		for li, line := range p.lines {
			if used[li] {
				continue
			}
			distance := line.Vectors[0].SqDist(*optimizedpos)
			if distance < bestdistance || bestdistance == -1 {
				bestdistance = distance
				bestindex = li
				bestisreversed = false
			}
			if reversible {
				distance = line.Vectors[len(line.Vectors)-1].SqDist(*optimizedpos)
				if distance < bestdistance {
					bestdistance = distance
					bestindex = li
					bestisreversed = true
				}
			}
		}

		if bestindex == -1 {
			panic("No line found")
		}

		bestline := p.lines[bestindex]
		if bestisreversed {
			bestline = bestline.Reverse()
		}

		if bestdistance <= 0.01 && len(neworder) > 0 {
			// Merge them
			newlines := Lines{
				append(neworder[len(neworder)-1].Vectors, bestline.Vectors[1:]...),
			}
			neworder[len(neworder)-1] = newlines
		} else {
			neworder = append(neworder, bestline)
		}

		used[bestindex] = true
		*optimizedpos = bestline.Vectors[len(bestline.Vectors)-1]
	}
	return Plot{neworder}
}

func (p *Plot) String() string {
	var result string
	for _, line := range p.lines {
		for i, vector := range line.Vectors {
			if i == 0 {
				result += "PU"
			} else if i == 1 {
				result += "PD"
			} else {
				result += ","
			}
			result += fmt.Sprintf("%v,%v", vector.X, vector.Y)
			if i == 0 {
				result += ";"
			}
		}
		result += ";"
	}
	return result
}

type SimpleCommand string

type Command interface {
	Command() string
}

func main() {
	inputfile := pflag.String("input", "", "input file")
	outputfile := pflag.String("output", "", "output file")
	reversible := pflag.Bool("reversible", true, "are we allowed to plot lines reverse")
	breakandassemble := pflag.Bool("breakandassemble", true, "break lines and reassemble them")
	pflag.Parse()

	hpglraw, err := ioutil.ReadFile(*inputfile)
	if err != nil {
		panic(err)
	}
	hpgl := string(hpglraw)

	hpgl = strings.ReplaceAll(hpgl, "\n", "")
	hpgl = strings.ReplaceAll(hpgl, "\n", "")
	hpgl = strings.ToUpper(hpgl)

	output, err := os.Create(*outputfile)

	var reorderable Plot
	var currentpos, optimizedpos geom.Vec2

	for _, command := range strings.Split(hpgl, ";") {
		if len(command) < 2 {
			continue
		}
		operation := command[0:2]
		var rest string
		if len(command) > 2 {
			rest = command[2:]
		}
		switch {
		case operation == "PD" && rest != "":
			var operation Lines
			valuecount := 0

			operation.Vectors = append(operation.Vectors, currentpos)

			for _, v := range strings.Split(rest, ",") {
				vi, err := strconv.ParseFloat(v, 32)
				if err != nil {
					panic(err)
				}
				if valuecount%2 == 0 {
					currentpos.X = float32(vi)
				} else {
					currentpos.Y = float32(vi)

					operation.Vectors = append(operation.Vectors, currentpos)

					if *breakandassemble {
						reorderable.lines = append(reorderable.lines, operation)
						operation = Lines{
							Vectors: []geom.Vec2{currentpos},
						}
					}
				}
				valuecount++
			}

			if !*breakandassemble {
				reorderable.lines = append(reorderable.lines, operation)
			}
		case operation == "PU" && rest != "":
			valuecount := 0
			for _, v := range strings.Split(rest, ",") {
				vi, err := strconv.ParseFloat(v, 32)
				if err != nil {
					panic(err)
				}
				if valuecount%2 == 0 {
					currentpos.X = float32(vi)
				} else {
					currentpos.Y = float32(vi)
				}
				valuecount++
				if valuecount > 2 {
					panic("Too many PU values")
				}
			}
		default:
			if len(reorderable.lines) > 0 {
				optimized := reorderable.Optimize(&optimizedpos, *reversible)
				fmt.Fprint(output, optimized.String())
				fmt.Printf("Optimized moves %v, unoptimized %v\n", optimized.MoveDistances(currentpos), reorderable.MoveDistances(currentpos))
				reorderable.lines = make([]Lines, 0)
				currentpos = optimizedpos
			}
			fmt.Fprintf(output, "%v;", command)
		}
	}
	if len(reorderable.lines) >= 0 {
		optimized := reorderable.Optimize(&optimizedpos, *reversible)
		fmt.Fprint(output, optimized.String())
		reorderable.lines = make([]Lines, 0)
	}
	output.Close()
}
