package id

// New can get a unique code by id(You need to ensure that id is unique)
func New(id uint64, options ...func(*Options)) string {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// enlarge and add salt
	id = id*uint64(ops.n1) + ops.salt

	var code []rune
	slIdx := make([]byte, ops.l)

	charLen := len(ops.chars)
	charLenUI := uint64(charLen)

	// diffusion
	for i := 0; i < ops.l; i++ {
		slIdx[i] = byte(id % charLenUI)                          // get each number
		slIdx[i] = (slIdx[i] + byte(i)*slIdx[0]) % byte(charLen) // let units digit affect other digit
		id = id / charLenUI                                      // right shift
	}

	// confusion(https://en.wikipedia.org/wiki/Permutation_box)
	for i := 0; i < ops.l; i++ {
		idx := (byte(i) * byte(ops.n2)) % byte(ops.l)
		code = append(code, ops.chars[slIdx[idx]])
	}
	return string(code)
}
