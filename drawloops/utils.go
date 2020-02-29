package drawloops

//Linspace function returns a slice of n values which are linearly spread out between a and b.
func linspace(a, b float64, n int) []float64 {

	//At least two points are required
	if n < 2 {
		return nil
	}

	//b value has to be bigger than a
	if b < a {
		return nil
	}

	//Step size
	c := (b - a) / float64(n-1)

	//Create and fill the slice
	out := make([]float64, n)
	for i := range out {
		out[i] = a + float64(i)*c
	}

	//Fix last entry to be b
	out[len(out)-1] = b

	return out
}

//calculateBarriers function returns the values from the result of the fourier transform
//that will be appropriate to a vertical bar level
func calculateBarriers(height int, minVal, maxVal float64) []float64 {
	space := linspace(minVal, maxVal, height)

	//reverse the order of the elements in the slice
	for left, right := 0, len(space)-1; left < right; left, right = left+1, right-1 {
		space[left], space[right] = space[right], space[left]
	}
	return space
}
