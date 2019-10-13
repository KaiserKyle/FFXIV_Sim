package main

import (
	"fmt"
	"math"
)

type enemy struct {
	Name         string
	MaxHP        int
	DamageDoneTo int
	Effects      []effect
	DoTEffects   []dotEffect
	TotalTime    float64
}

func (e *enemy) applyEffect(eff effect) {
	effectAlreadyApplied := false
	for i := range e.Effects {
		if e.Effects[i].Name == eff.Name {
			// Note: Do NOT reset TimeToNextTick
			effectAlreadyApplied = true
			e.Effects[i].Duration = eff.Duration
			e.Effects[i].TimeLeft = eff.TimeLeft
		}
	}

	if !effectAlreadyApplied {
		e.Effects = append(e.Effects, eff)
	}

	globalLog(Basic, fmt.Sprintf("    [EFFECT APPLIED] %s to %s: %v %v", eff.Name, e.Name, eff.OffensiveBuffs, eff.DefensiveBuffs))
}

func (e *enemy) applyDoTEffect(eff dotEffect) {
	dotAlreadyApplied := false
	for i := range e.DoTEffects {
		if e.DoTEffects[i].PlayerName == eff.PlayerName && e.DoTEffects[i].Name == eff.Name {
			// Note: Do NOT reset TimeToNextTick
			dotAlreadyApplied = true
			e.DoTEffects[i].Duration = eff.Duration
			e.DoTEffects[i].BaseDamage = eff.BaseDamage
			e.DoTEffects[i].CritChance = eff.CritChance
			e.DoTEffects[i].CritDamage = eff.CritDamage
			e.DoTEffects[i].DirectChance = eff.DirectChance
		}
	}

	if !dotAlreadyApplied {
		e.DoTEffects = append(e.DoTEffects, eff)
	}

	globalLog(Basic, fmt.Sprintf("    [DOT APPLIED] %s to %s [Base Damage:%d] [Rates:%.2f,%.2f]", eff.Name, e.Name, eff.BaseDamage, eff.CritChance, eff.DirectChance))
}

func (e *enemy) applyDamage(result *skillResult) {
	for i := range e.Effects {
		if len(e.Effects[i].DefensiveBuffs) > EnemyPiercing {
			buff := 100.0 - float64(e.Effects[i].DefensiveBuffs[EnemyPiercing])
			buff = buff / 100
			result.DamageBuff = result.DamageBuff * buff

			result.DamageDone = int(math.Floor(float64(result.DamageDone) * buff))
		}
	}
}

func (e *enemy) advanceTime(span float64) []skillResult {
	var results []skillResult

	e.TotalTime += span

	// Check for DoTs
	for i := range e.DoTEffects {
		// Advance the timer
		e.DoTEffects[i].TimeToNextTick -= span
		e.DoTEffects[i].Duration -= span
		// Check if we should apply a DoT
		if e.DoTEffects[i].TimeToNextTick <= 0.0 {
			// Apply DoT
			e.DoTEffects[i].TimeToNextTick = 3.0
			result := new(skillResult)

			result.SkillName = e.DoTEffects[i].Name
			result.PlayerName = e.DoTEffects[i].PlayerName
			result.TargetName = e.Name
			result.TimePerformed = e.TotalTime

			result.calculateDoTTick(e.DoTEffects[i])

			results = append(results, *result)
		}
	}

	// Iterate through buffs and remove expired ones
	buffCount := 0
	for j := range e.Effects {
		if e.Effects[j].TimeLeft > 0.0 {
			e.Effects[j].TimeLeft -= span
		}
		if e.Effects[j].TimeLeft > 0.0 {
			e.Effects[buffCount] = e.Effects[j]
			buffCount++
		}
	}

	e.Effects = e.Effects[:buffCount]

	buffCount = 0
	for k := range e.DoTEffects {
		if e.DoTEffects[k].Duration > 0.0 {
			e.DoTEffects[buffCount] = e.DoTEffects[k]
			buffCount++
		}
	}

	e.DoTEffects = e.DoTEffects[:buffCount]

	return results
}

func (e *enemy) reset() {
	e.DoTEffects = e.DoTEffects[:0]
	e.Effects = e.Effects[:0]
	e.TotalTime = 0
}
