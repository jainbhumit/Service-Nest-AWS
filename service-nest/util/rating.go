package util

func CalculateRating(avgRating float64, ratingCount int64, rating float64) float64 {
	var updatedRating float64
	totalReview := ratingCount + 1
	logic := (avgRating * float64(ratingCount)) + rating
	updatedRating = logic / float64(totalReview)
	return updatedRating
}
