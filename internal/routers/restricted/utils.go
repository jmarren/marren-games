package restricted

import (
	"fmt"
	"time"
)

// Function to compare sqlite date to httpFormat date
func CompareTimestampToHttp(lastModified time.Time, ifModifiedSince time.Time) (bool, error) {
	fmt.Println("lastModified: ", lastModified)
	fmt.Println("ifModifiedSince: ", ifModifiedSince)
	return false, nil
}
