package format_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/format"
)

type StringAlias string
type ByteAlias []byte
type IntAlias int

type AStruct struct {
	Exported string
}

type SimpleStruct struct {
	Name        string
	Enumeration int
	Veritas     bool
	Data        []byte
	secret      uint32
}

type ComplexStruct struct {
	Strings      []string
	SimpleThings []*SimpleStruct
	DataMaps     map[int]ByteAlias
}

var _ = Describe("Format", func() {
	match := func(typeRepresentation string, valueRepresentation string, args ...interface{}) OmegaMatcher {
		if len(args) > 0 {
			valueRepresentation = fmt.Sprintf(valueRepresentation, args...)
		}
		return Equal(fmt.Sprintf("%s<%s>: %s", Indent, typeRepresentation, valueRepresentation))
	}

	Describe("Message", func() {
		Context("with only an actual value", func() {
			It("should print out an indented formatted representation of the value and the message", func() {
				Ω(Message(3, "to be three.")).Should(Equal("Expected\n    <int>: 3\nto be three."))
			})
		})

		Context("with an actual and an expected value", func() {
			It("should print out an indented formatted representatino of both values, and the message", func() {
				Ω(Message(3, "to equal", 4)).Should(Equal("Expected\n    <int>: 3\nto equal\n    <int>: 4"))
			})
		})
	})

	Describe("IndentString", func() {
		It("should indent the string", func() {
			Ω(IndentString("foo\n  bar\nbaz", 2)).Should(Equal("        foo\n          bar\n        baz"))
		})
	})

	Describe("Object", func() {
		Describe("formatting boolean values", func() {
			It("should give the type and format values correctly", func() {
				Ω(Object(true, 1)).Should(match("bool", "true"))
				Ω(Object(false, 1)).Should(match("bool", "false"))
			})
		})

		Describe("formatting numbers", func() {
			It("should give the type and format values correctly", func() {
				Ω(Object(int(3), 1)).Should(match("int", "3"))
				Ω(Object(int8(3), 1)).Should(match("int8", "3"))
				Ω(Object(int16(3), 1)).Should(match("int16", "3"))
				Ω(Object(int32(3), 1)).Should(match("int32", "3"))
				Ω(Object(int64(3), 1)).Should(match("int64", "3"))

				Ω(Object(uint(3), 1)).Should(match("uint", "3"))
				Ω(Object(uint8(3), 1)).Should(match("uint8", "3"))
				Ω(Object(uint16(3), 1)).Should(match("uint16", "3"))
				Ω(Object(uint32(3), 1)).Should(match("uint32", "3"))
				Ω(Object(uint64(3), 1)).Should(match("uint64", "3"))
			})

			It("should handle uintptr differently", func() {
				Ω(Object(uintptr(3), 1)).Should(match("uintptr", "0x3"))
			})
		})

		Describe("formatting channels", func() {
			It("should give the type and format values correctly", func() {
				c := make(chan<- bool, 3)
				c <- true
				c <- false
				Ω(Object(c, 1)).Should(match("chan<- bool | len:2, cap:3", "%v", c))
			})
		})

		Describe("formatting strings", func() {
			It("should give the type and format values correctly", func() {
				s := "a\nb\nc"
				Ω(Object(s, 1)).Should(match("string", `a
    b
    c`))
			})
		})

		Describe("formatting []byte slices", func() {
			It("should present them as strings", func() {
				b := []byte("a\nb\nc")
				Ω(Object(b, 1)).Should(match("[]uint8 | len:5, cap:5", `a
    b
    c`))
			})
		})

		Describe("formatting functions", func() {
			It("should give the type and format values correctly", func() {
				f := func(a string, b []int) ([]byte, error) {
					return []byte("abc"), nil
				}
				Ω(Object(f, 1)).Should(match("func(string, []int) ([]uint8, error)", "%v", f))
			})
		})

		Describe("formatting pointers", func() {
			It("should give the type and dereference the value to format it correctly", func() {
				a := 3
				Ω(Object(&a, 1)).Should(match(fmt.Sprintf("*int | %p", &a), "3"))
			})

			Context("when there are pointers to pointers...", func() {
				It("should recursively deference the pointer until it gets to a value", func() {
					a := 3
					var b *int
					var c **int
					var d ***int
					b = &a
					c = &b
					d = &c

					Ω(Object(d, 1)).Should(match(fmt.Sprintf("***int | %p", d), "3"))
				})
			})

			Context("when the pointer points to nil", func() {
				It("should say nil and not explode", func() {
					var a *AStruct
					Ω(Object(a, 1)).Should(match("*format_test.AStruct | 0x0", "nil"))
				})
			})
		})

		Describe("formatting arrays", func() {
			It("should give the type and format values correctly", func() {
				w := [3]string{"Jed Bartlet", "Toby Ziegler", "CJ Cregg"}
				Ω(Object(w, 1)).Should(match("[3]string", `["Jed Bartlet", "Toby Ziegler", "CJ Cregg"]`))
			})
		})

		Describe("formatting slices", func() {
			It("should include the length and capacity in the type information", func() {
				s := make([]bool, 3, 4)
				Ω(Object(s, 1)).Should(match("[]bool | len:3, cap:4", "[false, false, false]"))
			})

			Context("when the slice contains long entries", func() {
				It("should format the entries with newlines", func() {
					w := []string{"Josiah Edward Bartlet", "Toby Ziegler", "CJ Cregg"}
					expected := `[
        "Josiah Edward Bartlet",
        "Toby Ziegler",
        "CJ Cregg",
    ]`
					Ω(Object(w, 1)).Should(match("[]string | len:3, cap:3", expected))
				})
			})
		})

		Describe("formatting maps", func() {
			It("should include the length in the type information", func() {
				m := make(map[int]bool, 5)
				m[3] = true
				m[4] = false
				Ω(Object(m, 1)).Should(match("map[int]bool | len:2", "{3: true, 4: false}"))
			})

			Context("when the slice contains long entries", func() {
				It("should format the entries with newlines", func() {
					m := map[string][]byte{}
					m["Josiah Edward Bartlet"] = []byte("Martin Sheen")
					m["Toby Ziegler"] = []byte("Richard Schiff")
					m["CJ Cregg"] = []byte("Allison Janney")
					expected := `{
        "Josiah Edward Bartlet": "Martin Sheen",
        "Toby Ziegler": "Richard Schiff",
        "CJ Cregg": "Allison Janney",
    }`
					Ω(Object(m, 1)).Should(match("map[string][]uint8 | len:3", expected))
				})
			})
		})

		Describe("formatting structs", func() {
			It("should include the struct name and the field names", func() {
				s := SimpleStruct{
					Name:        "Oswald",
					Enumeration: 17,
					Veritas:     true,
					Data:        []byte("datum"),
					secret:      1983,
				}

				Ω(Object(s, 1)).Should(match("format_test.SimpleStruct", `{Name: "Oswald", Enumeration: 17, Veritas: true, Data: "datum"}`))
			})

			Context("when the struct contains long entries", func() {
				It("should format the entries with new lines", func() {
					s := &SimpleStruct{
						Name:        "Mithrandir Gandalf Greyhame",
						Enumeration: 2021,
						Veritas:     true,
						Data:        []byte("wizard"),
						secret:      3,
					}

					Ω(Object(s, 1)).Should(match(fmt.Sprintf("*format_test.SimpleStruct | %p", s), `{
        Name: "Mithrandir Gandalf Greyhame",
        Enumeration: 2021,
        Veritas: true,
        Data: "wizard",
    }`))
				})
			})
		})

		Describe("formatting nil values", func() {
			It("should print out nil", func() {
				Ω(Object(nil, 1)).Should(match("nil", "nil"))
				var typedNil *AStruct
				Ω(Object(typedNil, 1)).Should(match("*format_test.AStruct | 0x0", "nil"))
				var c chan<- bool
				Ω(Object(c, 1)).Should(match("chan<- bool | len:0, cap:0", "nil"))
				var s []string
				Ω(Object(s, 1)).Should(match("[]string | len:0, cap:0", "nil"))
				var m map[string]bool
				Ω(Object(m, 1)).Should(match("map[string]bool | len:0", "nil"))
			})
		})

		Describe("formatting aliased types", func() {
			It("should print out the correct alias type", func() {
				Ω(Object(StringAlias("alias"), 1)).Should(match("format_test.StringAlias", `alias`))
				Ω(Object(ByteAlias("alias"), 1)).Should(match("format_test.ByteAlias | len:5, cap:5", `alias`))
				Ω(Object(IntAlias(3), 1)).Should(match("format_test.IntAlias", "3"))
			})
		})

		Describe("handling nested things", func() {
			It("should produce a correctly nested representation", func() {
				s := ComplexStruct{
					Strings: []string{"lots", "of", "short", "strings"},
					SimpleThings: []*SimpleStruct{
						{"short", 7, true, []byte("succinct"), 17},
						{"something longer", 427, true, []byte("designed to wrap around nicely"), 30},
					},
					DataMaps: map[int]ByteAlias{
						17:   ByteAlias("some substantially longer chunks of data"),
						1138: ByteAlias("that should make things wrap"),
					},
				}
				expected := `{
        Strings: ["lots", "of", "short", "strings"],
        SimpleThings: [
            {Name: "short", Enumeration: 7, Veritas: true, Data: "succinct"},
            {
                Name: "something longer",
                Enumeration: 427,
                Veritas: true,
                Data: "designed to wrap around nicely",
            },
        ],
        DataMaps: {
            17: "some substantially longer chunks of data",
            1138: "that should make things wrap",
        },
    }`
				Ω(Object(s, 1)).Should(match("format_test.ComplexStruct", expected))
			})
		})
	})
})
