package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// InGraveyardOfTargetsController is "in that player's graveyard",
// where slot names the earlier target permanent controlled by that player.
func InGraveyardOfTargetsController(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetInControllersGraveyard}
	return spec
}

// AttachedToTarget is "attached to that creature".
func AttachedToTarget(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetAttachedTo}
	return spec
}

// ControlledByTargetsController is "that player controls", where slot
// names the earlier target permanent controlled by that player
// (Breaking of the Fellowship's "another target creature that player
// controls"). If the earlier target has left the battlefield by
// resolution, its last-known controller still decides (CR 608.2h).
func ControlledByTargetsController(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetControlledBySameController}
	return spec
}

// SharesACardTypeWithTarget is "that shares a card type with it", where slot
// names the earlier target permanent (Burglar's Plot's "two target nonland
// permanents that share a card type"). It is judged as the spell is cast
// (CR 601.2c) and again as it resolves (CR 608.2b), on the permanents' types
// after the layers.
func SharesACardTypeWithTarget(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetSharesCardType}
	return spec
}

// SharesACreatureTypeWithTarget is "that share a creature type", where slot
// names the earlier target permanent (Secret Tunnel's "two target creatures
// you control that share a creature type"). It is judged as the ability is
// activated (CR 601.2c) and again as it resolves (CR 608.2b), on creature
// types after the layers; an earlier target that has left by then answers
// with the creature types it last had (CR 608.2h).
func SharesACreatureTypeWithTarget(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetSharesCreatureType}
	return spec
}

// SharesACreatureTypeWithTargetCard is SharesACreatureTypeWithTarget for two
// cards in a graveyard (Unbury's "two target creature cards that share a
// creature type"): the earlier pick is read wherever its card is now, so
// one that left the graveyard still anchors the other (the ruling).
func SharesACreatureTypeWithTargetCard(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetCardSharesCreatureType}
	return spec
}
