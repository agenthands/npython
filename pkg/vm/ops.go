package vm

const (
	OP_HALT      uint8 = 0x00
	OP_NOOP      uint8 = 0x01
	OP_PUSH_C    uint8 = 0x02
	OP_PUSH_L    uint8 = 0x03
	OP_POP_L     uint8 = 0x04
	OP_DUP       uint8 = 0x05
	OP_ADD       uint8 = 0x10
	OP_SUB       uint8 = 0x11
	OP_MUL       uint8 = 0x12
	OP_DIV       uint8 = 0x1a
	OP_EQ        uint8 = 0x13
	OP_NE        uint8 = 0x1b
	OP_GT        uint8 = 0x14
	OP_LT        uint8 = 0x18
	OP_DROP      uint8 = 0x1c
	OP_PRINT     uint8 = 0x15
	OP_CONTAINS  uint8 = 0x16
	OP_FIND      uint8 = 0x1d
	OP_SLICE     uint8 = 0x1e
	OP_LEN       uint8 = 0x1f
	OP_TRIM      uint8 = 0x24
	OP_MOD       uint8 = 0x25
	OP_LTE       uint8 = 0x26
	OP_GTE       uint8 = 0x27
	OP_POW       uint8 = 0x28
	OP_AND       uint8 = 0x32
	OP_OR        uint8 = 0x33
	OP_IN        uint8 = 0x34
	OP_NOT_IN    uint8 = 0x35
	OP_ERROR     uint8 = 0x17
	OP_JMP       uint8 = 0x20
	OP_JMP_FALSE uint8 = 0x21
	OP_CALL      uint8 = 0x22
	OP_RET       uint8 = 0x23
	OP_FLOOR_DIV uint8 = 0x29
	OP_BIT_AND   uint8 = 0x2a
	OP_BIT_OR    uint8 = 0x2b
	OP_BIT_XOR   uint8 = 0x2c
	OP_LSHIFT    uint8 = 0x2d
	OP_RSHIFT    uint8 = 0x2e
	OP_ADDRESS   uint8 = 0x30
	OP_EXIT_ADDR uint8 = 0x31
	OP_SYSCALL   uint8 = 0x40
)
