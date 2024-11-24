package betterfieldmask

import (
	"fmt"
	"strings"

	"github.com/mennanov/fmutils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Update[T proto.Message](old, update T, updateMask *fieldmaskpb.FieldMask) T {

	merged := proto.Clone(old)

	fmutils.Filter(update, updateMask.Paths)

	for _, p := range updateMask.Paths {
		fmt.Println(p)
		splt := strings.Split(p, ".")
		val, ok := lookupVal(splt, update.ProtoReflect())
		if !ok {
			continue
		}

		ok = setVal(splt, merged.ProtoReflect(), val)
		if !ok {
			continue
		}
	}

	return merged.(T) //nolint:revive // updated is guaranteed to be of type T, because we cloned it from old, which is of type T.
}

func setValMap(pathSegments []string, fd protoreflect.FieldDescriptor, msg protoreflect.Map, ov protoreflect.Value) bool {
	cur := pathSegments[0]
	pathSegments = pathSegments[1:]

	fmt.Println("segMap", pathSegments, "cur", cur)

	var result bool

	// TODO support all map types
	key := protoreflect.ValueOf(cur).MapKey()

	val := msg.Get(key)
	if !val.IsValid() {
		if len(pathSegments) == 0 {
			msg.Set(key, ov)
			return true
		} else {
			nw := msg.NewValue()
			msg.Set(key, nw)
			val = nw
		}
	}

	if fd.MapValue().Kind() == protoreflect.MessageKind {
		return setVal(pathSegments, val.Message(), ov)
	}

	return result
}

func setVal(pathSegments []string, msg protoreflect.Message, ov protoreflect.Value) bool {
	fmt.Println("CALL")
	cur := pathSegments[0]
	pathSegments = pathSegments[1:]

	fmt.Println("seg", pathSegments)

	var result bool

	fmt.Println("SV", msg.Descriptor().FullName(), "looking for", cur)

	// How to deal with unset fields?

	//  TODO If field is type msg, and we're not at the end, set it empty.

	fd := msg.Descriptor().Fields().ByName(protoreflect.Name(cur))

	if len(pathSegments) == 0 {
		// TODO make sure types match
		msg.Set(fd, ov)
		result = true
		return false
	} else {
		maybeVal := msg.Get(fd)

		var createNew bool
		switch fd.Kind() {
		case protoreflect.MessageKind:
			if fd.IsMap() {
				createNew = !maybeVal.Map().IsValid()
			} else {
				createNew = !maybeVal.Message().IsValid()
			}
		default:
		}

		// Pre-check: Field not exist. set it.
		if createNew {
			fmt.Println("!!!!!!!!!!!!!!!! SET")
			maybeVal = msg.NewField(fd)
			msg.Set(fd, maybeVal)

		}

		if fd.Kind() == protoreflect.MessageKind {
			// IS MAP??
			if fd.IsMap() {
				result = setValMap(pathSegments, fd, maybeVal.Map(), ov)
				return false
			} else {
				result = setVal(pathSegments, maybeVal.Message(), ov)
				return false
			}
		}

	}

	// msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	// 	fmt.Println("RANGE")
	// 	fmt.Println(fd.Name())
	// 	if string(fd.Name()) == cur {
	// 		fmt.Println("match", fd.Name(), cur)
	// 		if len(pathSegments) == 0 {
	// 			fmt.Println("TERM")
	// 			msg.Set(fd, ov)
	// 			result = true
	// 			return false
	// 		}
	//
	// 		// recurse
	// 		if fd.Kind() == protoreflect.MessageKind {
	// 			// IS MAP??
	// 			if fd.IsMap() {
	// 				result = setValMap(pathSegments, fd, v.Map(), ov)
	// 				return false
	// 			} else {
	// 				result = setVal(pathSegments, v.Message(), ov)
	// 				return false
	// 			}
	// 		}
	//
	// 		// Bug/wrong input
	//
	// 	}
	// 	return true
	// })
	//
	// if !result {
	// 	fd := msg.Descriptor().Fields().ByName(protoreflect.Name(cur))
	// 	if len(pathSegments) == 0 {
	// 		msg.Set(fd, ov)
	// 	} else {
	// 		// field is not set. set it.
	// 		fmt.Println("MUST SET PLZ", cur)
	//
	// 		// Create either
	//
	// 		// A) Message
	// 		// B) Map
	// 		// C) List
	//
	// 		// fd.Message().
	//
	// 		// How to create message?
	//
	// 		// x := protoreflect.ValueOfMessage(dynamicpb.NewMessage(fd.Message()))
	//
	// 		newVal := msg.NewField(fd)
	// 		msg.Set(fd, newVal)
	// 		// NOw, also recurse into it.
	//
	// 	}
	//
	// }

	return result
}

func lookupVal(pathSegments []string, msg protoreflect.Message) (protoreflect.Value, bool) {
	cur := pathSegments[0]
	pathSegments = pathSegments[1:]

	var result *protoreflect.Value

	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fmt.Println(fd.Name())
		if string(fd.Name()) == cur {
			fmt.Println("K")

			if len(pathSegments) == 0 {
				fmt.Println("terminal", cur)
				result = &v
				return false
			}

			// Recurse deeper

			// Handle kind List.

			if fd.Kind() == protoreflect.MessageKind {
				if fd.IsMap() {
					rez, ok := lookupValMap(pathSegments, fd, v.Map())
					if ok {
						result = &rez
					}
					return false

				} else {
					rez, ok := lookupVal(pathSegments, v.Message())
					if ok {
						result = &rez
					}
					return false
				}
			}

			// we could track that we found the field, but it's wrong type.

		}
		return true
	})

	if result == nil {
		return protoreflect.Value{}, false
	}

	return *result, true
}

func lookupValMap(pathSegments []string, fd protoreflect.FieldDescriptor, msg protoreflect.Map) (protoreflect.Value, bool) {
	cur := pathSegments[0]
	_ = cur
	pathSegments = pathSegments[1:]
	var result *protoreflect.Value

	msg.Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
		k := mk.String()
		fmt.Println("mapkey", k)

		if k == cur {
			fmt.Println("OK")

			// Check if paths empty. If yes, return it and be done!
			if len(pathSegments) == 0 {
				fmt.Println("terminal")
				result = &v
				return false
			}

			if fd.MapValue().Kind() == protoreflect.MessageKind {
				// map value can't be maps themselves. they are always message, or primitive
				rez, ok := lookupVal(pathSegments, v.Message())
				if ok {
					result = &rez
				}
				return false
			}

			// TODO handle list.

			// is primitive , but segments are left. error.

		}

		return true
	})

	if result == nil {
		return protoreflect.Value{}, false
	}

	return *result, true
}

func mergeverse(path string, msg protoreflect.Message, data protoreflect.Value) {
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fmt.Println(fd.Name())
		nestedPath := path + string(fd.Name()) + "."
		// We only recurse into messages. Lists (repeated) and Maps are added
		// "completely" to the fieldmask.
		if fd.Kind() == protoreflect.MessageKind && !fd.IsMap() && !fd.IsList() {
			mergeverse(nestedPath, v.Message(), data)
			return true
		}
		return true
	})
}
