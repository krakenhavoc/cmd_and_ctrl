package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cranial Plating — Artifact — Equipment for {2} (EDHREC rank 2421):
//
//	"Equipped creature gets +1/+0 for each artifact you control.
//	 {B}{B}: Attach this Equipment to target creature you control.
//	 Equip {1}"
//
// The artifact deck's kill condition, and the card that shows the
// second half of ADR 0036 decision 4: if equip is an ordinary
// activated ability, then an ability that attaches WITHOUT being an
// equip ability is the same entry minus the sorcery-speed gate.
//
// That missing gate is the whole card. {B}{B} at instant speed means
// the Plating moves in response to removal, moves onto an unblocked
// creature after blockers are declared, and moves onto a second
// attacker mid-combat. Every one of those is a real play and every
// one of them falls out of `SorcerySpeed: false` — the target clause,
// the announce check, the CR 608.2b re-check and the attach are all
// the ones the equip ability already uses.
//
// The bonus reuses artifactsControlledBy, the same counter
// Storm-Kiln Artist reads, so both cards get the same answer about
// what an artifact is: post-layer types, which means an animated
// permanent counts. The Plating counts ITSELF, which is correct — it
// is an artifact you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75564721-e4e9-463a-b49e-f0c7cd6f53a7",
		Name:         "Cranial Plating",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, effects.ArtifactsControlledBy),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}{B}: Attach to target creature you control",
				Cost:    effects.ManaCost("{B}{B}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.AttachSourceToTarget,
			},
			effects.EquipAbility("{1}"),
		},
	})
}
