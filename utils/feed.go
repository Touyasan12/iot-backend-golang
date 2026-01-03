package utils

const DefaultFeedDoseGram = 10

// NormalizeFeedAmount ensures the feed amount is at least the default dose.
func NormalizeFeedAmount(amountGram int) int {
	if amountGram <= 0 {
		return DefaultFeedDoseGram
	}
	return amountGram
}

// CalculateFeedDoses converts gram amount to number of doses (ceil division).
func CalculateFeedDoses(amountGram int) int {
	amount := NormalizeFeedAmount(amountGram)
	doses := amount / DefaultFeedDoseGram
	if amount%DefaultFeedDoseGram != 0 {
		doses++
	}
	if doses < 1 {
		doses = 1
	}
	return doses
}

