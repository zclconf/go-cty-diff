package cty

import (
	"sort"

	"github.com/apparentlymart/go-cty/cty/set"
)

// ElementIterator is the interface type returned by Value.ElementIterator to
// allow the caller to iterate over elements of a collection-typed value.
//
// Its usage pattern is as follows:
//
//     it := val.ElementIterator()
//     for it.Next() {
//         key, val := it.Element()
//         // ...
//     }
type ElementIterator interface {
	Next() bool
	Element() (key Value, value Value)
}

func elementIterator(val Value) ElementIterator {
	switch {
	case val.ty.IsListType():
		return &listElementIterator{
			ety:  val.ty.ElementType(),
			vals: val.v.([]interface{}),
			idx:  -1,
		}
	case val.ty.IsMapType():
		// We iterate the keys in a predictable lexicographical order so
		// that results will always be stable given the same input map.
		rawMap := val.v.(map[string]interface{})
		keys := make([]string, 0, len(rawMap))
		for key := range rawMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		return &mapElementIterator{
			ety:  val.ty.ElementType(),
			vals: rawMap,
			keys: keys,
			idx:  -1,
		}
	case val.ty.IsSetType():
		rawSet := val.v.(set.Set)
		return &setElementIterator{
			ety:   val.ty.ElementType(),
			setIt: rawSet.Iterator(),
		}
	case val.ty.IsTupleType():
		return &tupleElementIterator{
			etys: val.ty.TupleElementTypes(),
			vals: val.v.([]interface{}),
			idx:  -1,
		}
	default:
		panic("attempt to iterate on non-collection, non-tuple type")
	}
}

type listElementIterator struct {
	ety  Type
	vals []interface{}
	idx  int
}

func (it *listElementIterator) Element() (Value, Value) {
	i := it.idx
	return NumberIntVal(int64(i)), Value{
		ty: it.ety,
		v:  it.vals[i],
	}
}

func (it *listElementIterator) Next() bool {
	it.idx++
	return it.idx < len(it.vals)
}

type mapElementIterator struct {
	ety  Type
	vals map[string]interface{}
	keys []string
	idx  int
}

func (it *mapElementIterator) Element() (Value, Value) {
	key := it.keys[it.idx]
	return StringVal(key), Value{
		ty: it.ety,
		v:  it.vals[key],
	}
}

func (it *mapElementIterator) Next() bool {
	it.idx++
	return it.idx < len(it.keys)
}

type setElementIterator struct {
	ety   Type
	setIt *set.Iterator
}

func (it *setElementIterator) Element() (Value, Value) {
	val := Value{
		ty: it.ety,
		v:  it.setIt.Value(),
	}
	return val, val
}

func (it *setElementIterator) Next() bool {
	return it.setIt.Next()
}

type tupleElementIterator struct {
	etys []Type
	vals []interface{}
	idx  int
}

func (it *tupleElementIterator) Element() (Value, Value) {
	i := it.idx
	return NumberIntVal(int64(i)), Value{
		ty: it.etys[i],
		v:  it.vals[i],
	}
}

func (it *tupleElementIterator) Next() bool {
	it.idx++
	return it.idx < len(it.vals)
}
