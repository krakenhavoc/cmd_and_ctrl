package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zetalpa, Primal Dawn — Legendary Creature — Elder Dinosaur
// {6}{W}{W}, 4/8 (Foundations, #4465):
//
//	"Flying, double strike, vigilance, trample, indestructible"
//
// Five printed keywords the engine enforces from PrintedKeywords. The
// legend rule rides the type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7da0e5de-3e4c-420a-8685-991206100b9d",
		Name:            "Zetalpa, Primal Dawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike", "vigilance", "trample", "indestructible"},
	})
}
