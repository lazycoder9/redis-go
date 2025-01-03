package main

func Chunk[Slice ~[]E, E any](s Slice, n int) []Slice {
	if n < 1 {
		panic("cannot be less than 1")
	}

	result := make([]Slice, 0, (len(s)+n-1)/n)

	for i := 0; i < len(s); i += n {
		// Clamp the last chunk to the slice bound as necessary.
		end := min(n, len(s[i:]))

		// Set the capacity of each chunk so that appending to a chunk does
		// not modify the original slice.
		result = append(result, s[i:i+end:i+end])
	}

	return result
}
