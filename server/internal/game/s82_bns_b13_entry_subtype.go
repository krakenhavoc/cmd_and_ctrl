package game

// AddSubtypeAtETB makes the entering permanent a `subtype` IN ADDITION to
// its other types, as a replacement effect on its entry (CR 614.1c, CR
// 205.3d): Master Biomancer's "...and as a Mutant in addition to its other
// types". Only meaningful on a RepEventMove to the battlefield.
//
// It rides the same entry overwrite a putting effect's "it's a <type>"
// uses (ADR 0210), registered before the zone-move event and EventETB, so
// nothing sees the object without the subtype. The overwrite is copied
// before it is written, because the batch's own options share the pointer.
func (ev *ReplacementEvent) AddSubtypeAtETB(subtype string, label string) {
	if ev == nil || subtype == "" || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield {
		return
	}
	e := copyEntersAs(ev.entersAs)
	if e == nil {
		e = &EntersAs{Label: label}
	}
	if e.Label == "" {
		e.Label = label
	}
	if len(e.Types) > 0 {
		e.Subtypes = append(e.Subtypes, subtype)
	} else {
		e.AddSubtypes = append(e.AddSubtypes, subtype)
	}
	ev.entersAs = e
}
