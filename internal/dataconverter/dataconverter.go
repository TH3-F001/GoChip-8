package dataconverter

var boolByteMap map[bool]byte = map[bool]byte{
	true:  1,
	false: 0,
}

var byteBoolMap map[byte]bool = map[byte]bool{
	1: true,
	0: false,
}

var byteRuneMap map[byte]rune = map[byte]rune{
	1: '░',
	0: ' ',
}

var boolRuneMap map[bool]rune = map[bool]rune{
	true:  '░',
	false: ' ',
}

var runeBoolMap map[rune]bool = map[rune]bool{
	'░': true,
	' ': false,
}

var intByteMap map[int]byte = map[int]byte{
	1: 1,
	0: 0,
}

var byteIntMap map[byte]int = map[byte]int{
	1: 1,
	0: 0,
}

var uintByteMap map[uint]byte = map[uint]byte{
	1: 1,
	0: 0,
}

var byteUintMap map[byte]uint = map[byte]uint{
	1: 1,
	0: 0,
}

// BoolToByte ... converts a boolean to its corresponding byte value using a predefined map.
func BoolToByte(val bool) byte {
	return boolByteMap[val]
}

// ByteToBool ... converts a byte to its corresponding boolean value using a predefined map.
func ByteToBool(b byte) bool {
	return byteBoolMap[b%2]
}

// ByteToRune ... converts a byte to its corresponding rune value using a predefined map.
func ByteToRune(b byte) rune {
	return byteRuneMap[b%2]
}

// BoolToRune ... converts a boolean to its corresponding rune value using a predefined map.
func BoolToRune(val bool) rune {
	return boolRuneMap[val]
}

// IntToByte ... converts an integer to its corresponding byte value using a predefined map.
func IntToByte(i int) byte {
	return intByteMap[i%2]
}

// ByteToInt ... converts a byte to its corresponding integer value using a predefined map.
func ByteToInt(b byte) int {
	return byteIntMap[b%2]
}

// UintToByte ... converts an integer to its corresponding byte value using a predefined map.
func UintToByte(u uint) byte {
	return uintByteMap[u%2]
}

// ByteToInt ... converts a byte to its corresponding integer value using a predefined map.
func ByteToUint(b byte) uint {
	return byteUintMap[b%2]
}
