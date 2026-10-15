package effects

import "fmt"

// checkCastTimeValues refuses a CastTimeValues declaration the stamp
// could not honour (#3781): a blank or duplicated key is unaddressable
// from ctx.CastValue, and a nil Value would stamp nothing while the
// resolution waits for it. Copy-paste mistakes, failed at boot.
func checkCastTimeValues(spec Spec) {
	seen := make(map[string]bool, len(spec.CastTimeValues))
	for _, v := range spec.CastTimeValues {
		if v.Key == "" {
			panic(fmt.Sprintf("effects.Register: %q declares a cast-time value with no Key", spec.Name))
		}
		if v.Value == nil {
			panic(fmt.Sprintf("effects.Register: %q declares cast-time value %q with no Value", spec.Name, v.Key))
		}
		if seen[v.Key] {
			panic(fmt.Sprintf("effects.Register: %q declares two cast-time values keyed %q", spec.Name, v.Key))
		}
		seen[v.Key] = true
	}
}
