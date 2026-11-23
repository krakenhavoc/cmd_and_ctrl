package game

// BecomeCreatureAtETB makes the entering permanent a creature of `subtype`
// with base power and toughness `power`/`toughness`, IN ADDITION to its other
// types, as a replacement effect on its entry (CR 614.1c, CR 205.1b, layers 4
// and 7b): Displaced Dinosaurs' "As a historic permanent you control enters,
// it becomes a 7/7 Dinosaur creature in addition to its other types". Only
// meaningful on a RepEventMove to the battlefield.
//
// It rides the entry overwrite a putting effect's "it's a <type>" uses (ADR
// 0210), registered before the zone-move event and EventETB, so nothing sees
// the object without the new characteristics. The overwrite is copied before
// it is written, because the batch's own options share the pointer.
func (ev *ReplacementEvent) BecomeCreatureAtETB(subtype string, power, toughness int, label string) {
	if ev == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield {
		return
	}
	e := copyEntersAs(ev.entersAs)
	if e == nil {
		e = &EntersAs{Label: label}
	}
	if e.Label == "" {
		e.Label = label
	}
	e.AddTypes = append(e.AddTypes, "Creature")
	if subtype != "" {
		e.AddSubtypes = append(e.AddSubtypes, subtype)
	}
	e.BasePT = &EntersAsPT{Power: power, Toughness: toughness}
	ev.entersAs = e
}
