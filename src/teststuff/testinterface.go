package main

import (
	"log"
)

type interface1 interface {
	one()
	two()
	three()
}

type foo1 struct {
}

func (l foo1) one() {
	log.Printf("1")
}

func (l foo1) two() {
	log.Printf("2")
}

func (l foo1) three() {
	log.Printf("3")
}

type interface2 interface {
	one()
	two()
	three()
}

type foo2 struct {
}

func (l *foo2) one() {
	log.Printf("11")
}

func (l *foo2) two() {
	log.Printf("22")
}

func (l *foo2) three() {
	log.Printf("33")
}

func (l *foo2) four() {
	log.Printf("44")
}

func Foo() interface1 {
	var l interface1
	l = foo1{}
	return l
}

func Foo2() *foo2 {

	return &foo2{}
}

type foo3 struct {
	x int
}

func (z foo3) inc() {
	z.x = z.x + 1
}

func (z *foo3) incp() {
	z.x = z.x + 1
}

func (l *foo3) blah() {
	log.Printf("blah")
}

func main() {

	x := Foo()
	x.one()

	f, ok := x.(foo1)
	if ok {
		log.Printf("supported interface - %q\n", f.one())
	} else {
		log.Printf("not supported interface \n")
	}

}
