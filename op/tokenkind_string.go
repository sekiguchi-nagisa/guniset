// Code generated by "stringer -type TokenKind -trimprefix Token -linecomment"; DO NOT EDIT.

package op

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenId-0]
	_ = x[TokenRune-1]
	_ = x[TokenColon-2]
	_ = x[TokenComma-3]
	_ = x[TokenLParen-4]
	_ = x[TokenRParen-5]
	_ = x[TokenPlus-6]
	_ = x[TokenMinus-7]
	_ = x[TokenRange-8]
	_ = x[TokenSpace-9]
}

const _TokenKind_name = "identifiercodePoint:,()+-..space"

var _TokenKind_index = [...]uint8{0, 10, 19, 20, 21, 22, 23, 24, 25, 27, 32}

func (i TokenKind) String() string {
	if i < 0 || i >= TokenKind(len(_TokenKind_index)-1) {
		return "TokenKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenKind_name[_TokenKind_index[i]:_TokenKind_index[i+1]]
}
