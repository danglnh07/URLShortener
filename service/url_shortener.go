package service

const (
	base62chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base        = 62
)

// Encode a number into Base62
func EncodeBase62(num int64) string {
	/*
	 * Algorithm explains: just like with other base conversion, we repeatedly divide the number
	 * to 62 and append the characters at [remainder] position in the characters set
	 */

	if num == 0 {
		return string(base62chars[0])
	}

	result := ""
	for num > 0 {
		remainder := num % base
		result = string(base62chars[remainder]) + result
		num = num / base
	}
	return result
}

// Decode Base62 string back to number
func DecodeBase62(str string) int64 {
	var num int64
	for _, c := range str {
		num *= base
		switch {
		case '0' <= c && c <= '9':
			num += int64(c - '0')
		case 'A' <= c && c <= 'Z':
			num += int64(c-'A') + 10
		case 'a' <= c && c <= 'z':
			num += int64(c-'a') + 36
		}
	}

	return num
}
