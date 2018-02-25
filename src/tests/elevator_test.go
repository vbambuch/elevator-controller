package tests

import (
	C "consts"
	"testing"
	"helper"
	"bytes"
)

type byteTable struct { a byte; b byte; c byte; d byte; r []byte }

func TestGetInstr(t *testing.T)  {
	table := []byteTable{
		{C.MotorDirection, C.MotorUp, C.EmptyByte, C.EmptyByte, []byte{1, 100, 0, 0}},
		{C.OrderButtonLight, C.CabButton, byte(3), C.TurnOn, []byte{2, 2, 3, 1}},
	}

	for _, value := range table {
		instr := helper.GetInstruction(value.a, value.b, value.c, value.d)
		if bytes.Compare(instr, value.r) != 0 {
			t.Errorf("Incorrect instruction, got: %d, want: %d.", instr, value.r)
		}
	}
}

