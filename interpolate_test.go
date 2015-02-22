package dat

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkInterpolate(b *testing.B) {
	// Do some allocations outside the loop so they don't affect the results
	argEq1 := Eq{"f": 2, "x": "hi"}
	argEq2 := map[string]interface{}{"g": 3}
	argEq3 := Eq{"h": []int{1, 2, 3}}
	sq, args := Select("a", "b", "z", "y", "x").
		Distinct().
		From("c").
		Where("d = $1 OR e = $2", 1, "wat").
		Where(argEq1).
		Where(argEq2).
		Where(argEq3).
		GroupBy("i").
		GroupBy("ii").
		GroupBy("iii").
		Having("j = k").
		Having("jj = $1", 1).
		Having("jjj = $1", 2).
		OrderBy("l").
		OrderBy("l").
		OrderBy("l").
		Limit(7).
		Offset(8).
		ToSQL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Interpolate(sq, args)
	}
}

func TestInterpolateNil(t *testing.T) {
	args := []interface{}{nil}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = NULL")
}

func TestInterpolateInts(t *testing.T) {
	args := []interface{}{
		int(1),
		int8(-2),
		int16(3),
		int32(4),
		int64(5),
		uint(6),
		uint8(7),
		uint16(8),
		uint32(9),
		uint64(10),
	}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2 AND c = $3 AND d = $4 AND e = $5 AND f = $6 AND g = $7 AND h = $8 AND i = $9 AND j = $10", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = -2 AND c = 3 AND d = 4 AND e = 5 AND f = 6 AND g = 7 AND h = 8 AND i = 9 AND j = 10")
}

func TestInterpolateBools(t *testing.T) {
	args := []interface{}{true, false}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = 0")
}

func TestInterpolateFloats(t *testing.T) {
	args := []interface{}{float32(0.15625), float64(3.14159)}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 0.15625 AND b = 3.14159")
}

func TestInterpolateEscapeStrings(t *testing.T) {
	args := []interface{}{"hello", "\"pg's world\" \\\b\f\n\r\t\x1a"}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	// E'' is postgres-specific
	assert.Equal(t, "SELECT * FROM x WHERE a = 'hello' AND b = '\"pg''s world\" \\\b\f\n\r\t\x1a'", str)
}

func TestInterpolateSlices(t *testing.T) {
	args := []interface{}{[]int{1}, []int{1, 2, 3}, []uint32{5, 6, 7}, []string{"wat", "ok"}}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2 AND c = $3 AND d = $4", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = (1) AND b = (1,2,3) AND c = (5,6,7) AND d = ('wat','ok')")
}

type myString struct {
	Present bool
	Val     string
}

func (m myString) Value() (driver.Value, error) {
	if m.Present {
		return m.Val, nil
	}
	return nil, nil
}

func TestIntepolatingValuers(t *testing.T) {
	args := []interface{}{myString{true, "wat"}, myString{false, "fry"}}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 'wat' AND b = NULL")
}

func TestInterpolatingUnsafeStrings(t *testing.T) {
	args := []interface{}{NOW, DEFAULT, UnsafeString(`hstore`)}
	str, err := Interpolate("SELECT * FROM x WHERE one=$1 AND two=$2 AND three=$3", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE one=NOW() AND two=DEFAULT AND three=hstore")
}

func TestInterpolatingPointers(t *testing.T) {
	var one int32 = 1000
	var two int64 = 2000
	var three float32 = 3
	var four float64 = 4
	var five = "five"
	var six = true

	args := []interface{}{&one, &two, &three, &four, &five, &six}
	str, err := Interpolate("SELECT * FROM x WHERE one=$1 AND two=$2 AND three=$3 AND four=$4 AND five=$5 AND six=$6", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE one=1000 AND two=2000 AND three=3 AND four=4 AND five='five' AND six=1")
}

func TestInterpolatingNulls(t *testing.T) {
	var one *int32
	var two *int64
	var three *float32
	var four *float64
	var five *string
	var six *bool

	args := []interface{}{one, two, three, four, five, six}
	str, err := Interpolate("SELECT * FROM x WHERE one=$1 AND two=$2 AND three=$3 AND four=$4 AND five=$5 AND six=$6", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE one=NULL AND two=NULL AND three=NULL AND four=NULL AND five=NULL AND six=NULL")
}

func TestInterpolatingTime(t *testing.T) {
	var ptim *time.Time
	tim2 := time.Date(2004, time.January, 1, 1, 1, 1, 1, time.UTC)
	tim := time.Time{}

	args := []interface{}{ptim, tim, &tim2}

	str, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2 AND c = $3", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = NULL AND b = '0001-01-01 00:00:00' AND c = '2004-01-01 01:01:01'")
}

func TestInterpolateErrors(t *testing.T) {
	_, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)

	// no harm, no foul
	_, err = Interpolate("SELECT * FROM x WHERE", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)

	_, err = Interpolate("SELECT * FROM x WHERE a = $1", []interface{}{string([]byte{0x34, 0xFF, 0xFE})})
	assert.Equal(t, err, ErrNotUTF8)

	_, err = Interpolate("SELECT * FROM x WHERE a = $1", []interface{}{struct{}{}})
	assert.Equal(t, err, ErrInvalidValue)

	_, err = Interpolate("SELECT * FROM x WHERE a = $1", []interface{}{[]struct{}{{}, {}}})
	assert.Equal(t, err, ErrInvalidSliceValue)
}
