package vm

const (
	OP_HALT       uint8 = 0x00
	OP_NOOP       uint8 = 0x01
	OP_PUSH_C     uint8 = 0x02
	OP_PUSH_L     uint8 = 0x03
	OP_POP_L      uint8 = 0x04
	OP_ADD        uint8 = 0x10
	OP_SUB        uint8 = 0x11
	OP_MUL        uint8 = 0x12
	OP_EQ         uint8 = 0x13
	OP_GT         uint8 = 0x14
	OP_PRINT      uint8 = 0x15
	OP_CONTAINS   uint8 = 0x16
	OP_ERROR      uint8 = 0x17
	OP_JMP        uint8 = 0x20
	OP_JMP_FALSE  uint8 = 0x21
	OP_CALL       uint8 = 0x22
	OP_RET        uint8 = 0x23
	OP_ADDRESS    uint8 = 0x30
	OP_EXIT_ADDR  uint8 = 0x31
	OP_SYSCALL    uint8 = 0x40
)
