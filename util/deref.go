package util

func DerefString(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}

func DerefInt(p *int, fallback int) int {
	if p != nil {
		return *p
	}
	return fallback
}

func DerefInt32(p *int32, fallback int32) int32 {
	if p != nil {
		return *p
	}
	return fallback
}
