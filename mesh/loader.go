package mesh

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var spaces = regexp.MustCompile("[ \t\r]+")

// Create a new mesh from a .obj file
func LoadObjFile(name string) (m *Mesh, err error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fmt.Println("load mesh from", name)
	return LoadObj(r)
}

// Crate a new mesh from data
func LoadObj(r io.Reader) (m *Mesh, err error) {
	var line string
	defer func() {
		if errPanic := recover(); errPanic != nil {
			err = fmt.Errorf("Error %s parsing line: %s", errPanic, line)
		}
	}()
	m = New()
	scanner := bufio.NewScanner(r)
	group := ""
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		rec := spaces.Split(line, -1)
		switch rec[0] {
		case "v":
			data := parseF32(rec[1:4])
			m.AddVertex(data[0], data[1], data[2])
		case "vt":
			data := parseF32(rec[1:3])
			m.AddTexCoord(data[0], data[1])
		case "vn":
			data := parseF32(rec[1:4])
			m.AddNormal(data[0], data[1], data[2])
		case "f":
			if len(rec) == 4 {
				m.AddFace(parseEl(rec[1]), parseEl(rec[2]), parseEl(rec[3]))
			} else if len(rec) == 5 {
				m.AddFaceQuad(parseEl(rec[1]), parseEl(rec[2]), parseEl(rec[3]), parseEl(rec[4]))
			} else {
				err = fmt.Errorf("face with %d vertices not supported", len(rec)-1)
				return
			}
		case "g":
			if group != "" {
				m.Build()
			}
			if len(rec) > 1 {
				group = rec[1]
				fmt.Println("LoadObjFile: start group", group)
			}
		default:
			fmt.Printf("LoadObjFile: skip %s\n", line)
		}
	}
	err = scanner.Err()
	m.Build()
	m.Clear()
	return
}

func parseF32(in []string) []float32 {
	out := make([]float32, len(in))
	for i, str := range in {
		val, err := strconv.ParseFloat(str, 32)
		if err != nil {
			panic(err)
		}
		out[i] = float32(val)
	}
	return out
}

func parseEl(in string) El {
	var out [3]int
	slist := strings.Split(in, "/")
	for i, str := range slist {
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			panic(err)
		}
		out[i] = int(val)
	}
	return El{Vert: out[0], Tex: out[1], Norm: out[2]}
}
