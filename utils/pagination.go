package utils

import "math"

func Paginate(total, limit, offset int) (pages int, currentPage int, hasNext bool) {
	pages = int(math.Ceil(float64(total) / float64(limit)))
	currentPage = (offset / limit) + 1
	hasNext = offset+limit < total
	return
}
