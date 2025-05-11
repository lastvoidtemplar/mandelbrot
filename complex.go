package main

type Fraction struct {
	Numerator   int64
	Denominator int64
}

func gcd(a int64, b int64) int64 {
	var t int64
	for b != 0 {
		t = b
		b = a % b
		a = t
	}
	return a
}

func AddFranctions(a Fraction, b Fraction) Fraction {
	tempNumerator := a.Numerator*b.Denominator + b.Numerator*a.Denominator
	tempDenominator := a.Denominator * b.Denominator

	q := gcd(tempNumerator, tempDenominator)
	return Fraction{
		Numerator:   tempNumerator / q,
		Denominator: tempDenominator / q,
	}
}
