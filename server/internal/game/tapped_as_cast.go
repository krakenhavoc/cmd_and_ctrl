package game

// tapped_as_cast.go — #6168: a target restriction that applies only as
// the spell is cast.
//
//	"Enchant creature
//	 You can't choose an untapped creature as this spell's target as you
//	 cast it.
//	 You control enchanted creature." (Enthralling Hold)
//
// CR 303.4a: "An Aura spell requires a target, which is defined by its
// enchant ability." The enchant clause is also the Aura's ongoing
// legality condition (CR 303.4c, 704.5m), judged by the same TargetSpec
// (attachmentLegalLocked), so a plain "tapped creature" predicate would
// put the Aura into the graveyard the moment the creature untaps. The
// second sentence is a restriction on CHOOSING the target as the spell
// is cast (CR 601.2c), not part of the enchant ability, so it is read
// only on the casting walk: TargetSource.casting, which SourceCast sets
// for the card being cast and every later walk (the CR 608.2b re-check,
// the attachment check, a copy's targets, an Aura entering without
// being cast) leaves unset.

// tappedAsCastAdmits is TappedOnlyAsCast's candidate gate: true for a
// clause without it and for any walk but a cast's; on a cast's, the
// candidate must be tapped.
func (s *TargetSpec) tappedAsCastAdmits(src TargetSource, c Card) bool {
	return !s.TappedOnlyAsCast || !src.casting || c.Tapped
}
