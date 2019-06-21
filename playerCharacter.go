package main

import (
	"fmt"
	"math"
)

const level70Sub int = 364
const level70Div int = 2170
const level70Main int = 292
const level70DrgStr int = 115

const directBonus float64 = 1.2

const animationLock float64 = 0.75
const longAnimationLock float64 = 1.50

const autoAttackPot float64 = 110.0 / 100.0

type playerCharacter struct {
	Name                    string
	Class                   string
	WeaponDamage            int
	WeaponDelay             float64
	AttackPower             int
	CriticalHit             int
	DirectHit               int
	Determination           int
	SkillSpeed              int
	Skills                  []playerSkill
	Effects                 []effect
	LastWeaponskill         string
	IsGCDOnCooldown         bool
	GCDTimeRemaining        float64
	TotalTime               float64
	AnimationLock           float64
	AutoAttackTimeRemaining float64
	PlayerStance            stance
	RotationFile            string
}

type playerSkill struct {
	Name                    string
	Potency                 int
	IsWeaponskill           bool
	CooldownSeconds         int
	ComboAction             string
	ComboPotency            int
	AdditionalEffectName    string
	ComboEffect             string
	OnCooldown              bool
	CooldownTime            float64
	TimeRemainingOnCooldown float64
	LongAnimationLock       bool
	ActivateStance          bool
	CheckStance             bool
	AlertStance             bool
	ComboAlertStance        bool
	RequiredEffect          string
	RemoveRequiredEffect    bool
}

// Returns (damage done, onCooldown, didCrit, didDirect)
// If onCooldown is true, then no action was performed.
func (p *playerCharacter) performAction(skillName string, enemyData *enemy) *skillResult {
	result := new(skillResult)

	if p.AnimationLock > 0.0 {
		return nil
	}

	for i := range p.Skills {
		if p.Skills[i].Name == skillName {
			if !p.isSkillAvailable(skillName) {
				return nil
			}

			result.SkillName = skillName
			result.SkillIndex = i
			result.PlayerName = p.Name
			result.TargetName = enemyData.Name
			result.TimePerformed = p.TotalTime

			// Skip damage calculation for skills that are only buffs
			if p.Skills[i].Potency != 0 {
				result.calculateDamage(&p.Skills[i], *p, false, 0)

				if p.Skills[i].IsWeaponskill {
					p.LastWeaponskill = p.Skills[i].Name
					p.IsGCDOnCooldown = true
					p.GCDTimeRemaining = p.calculateGCD()
					p.AnimationLock = animationLock
				} else {
					p.Skills[i].OnCooldown = true
					p.Skills[i].TimeRemainingOnCooldown = p.Skills[i].CooldownTime
					if p.Skills[i].LongAnimationLock {
						p.AnimationLock = longAnimationLock
					} else {
						p.AnimationLock = animationLock
					}
				}
			}

			result.EffectApplied = append(result.EffectApplied, p.Skills[i].AdditionalEffectName)
			if result.IsCombo {
				result.EffectApplied = append(result.EffectApplied, p.Skills[i].ComboEffect)
			}
			stanceEffect := p.PlayerStance.processSkillExecution(result)
			if "" != stanceEffect {
				result.EffectApplied = append(result.EffectApplied, stanceEffect)
			}

			if p.Skills[i].ActivateStance {
				p.PlayerStance.activate()
			}

			// If the buff only lasts for one weaponskill, set its TimeLeft to zero here
			// This will cause the buff to be removed in the next frame
			if p.Skills[i].IsWeaponskill {
				for j := range p.Effects {
					if p.Effects[j].OnlyNextWeaponskill {
						p.Effects[j].TimeLeft = 0
					}
				}
			}

			// If a skill removes its required effect, set its time to zero
			if p.Skills[i].RemoveRequiredEffect {
				for j := range p.Effects {
					if p.Effects[j].Name == p.Skills[i].RequiredEffect {
						p.Effects[j].TimeLeft = 0
					}
				}
			}

			if p.Skills[i].AlertStance || (result.IsCombo && p.Skills[i].ComboAlertStance) {
				p.PlayerStance.alertSkill(skillName)
			}

			break
		}
	}

	return result
}

// Returns Auto-Attack if one is applied
func (p *playerCharacter) advanceTime(span float64, targettedEnemy *enemy) *skillResult {
	var result *skillResult

	p.TotalTime += span

	p.PlayerStance.advanceTime(span)

	// Advance cooldown timers and animation lock
	if 0.0 != p.GCDTimeRemaining {
		p.GCDTimeRemaining -= span
		if p.GCDTimeRemaining <= 0.0 {
			p.GCDTimeRemaining = 0.0
			p.IsGCDOnCooldown = false
		}
	}
	if 0.0 != p.AnimationLock {
		p.AnimationLock -= span
	}
	for i := range p.Skills {
		if p.Skills[i].TimeRemainingOnCooldown != 0.0 {
			p.Skills[i].TimeRemainingOnCooldown -= span
			if p.Skills[i].TimeRemainingOnCooldown <= 0.0 {
				p.Skills[i].TimeRemainingOnCooldown = 0.0
				p.Skills[i].OnCooldown = false
			}
		}
	}

	// Iterate through buffs and remove expired ones
	buffCount := 0
	for j := range p.Effects {
		if p.Effects[j].TimeLeft > 0.0 {
			p.Effects[j].TimeLeft -= span
		}
		if p.Effects[j].TimeLeft > 0.0 {
			p.Effects[buffCount] = p.Effects[j]
			buffCount++
		}
	}

	p.Effects = p.Effects[:buffCount]

	// Check auto attack
	p.AutoAttackTimeRemaining -= span
	if 0.0 >= p.AutoAttackTimeRemaining {
		p.AutoAttackTimeRemaining = p.WeaponDelay
		// Do auto attack
		result = new(skillResult)
		result.PlayerName = p.Name
		result.TargetName = targettedEnemy.Name
		result.SkillName = "Attack"
		result.TimePerformed = p.TotalTime
		result.calculateDamage(nil, *p, true, 0)
	}

	return result
}

// INCOMPLETE: Need to implement proper buffs here, mostly zeroes right now
func (p *playerCharacter) calculateGCD() float64 {
	msGCD := math.Floor((1000.0 - math.Floor(130.0*(float64(p.SkillSpeed)-float64(level70Sub))/float64(level70Div))) * float64(p.WeaponDelay*1000) / 1000.0)

	a := math.Floor(math.Floor(math.Floor((100.0-0.0)*(100.0-0.0)/100.0)*(100-0)/100) - 0)
	b := (0.0 - 100.0) / -100.0

	cGCD := math.Floor(math.Floor(math.Floor(math.Ceil(a*b)*msGCD/100.0)*100/1000) * 100 / 100)

	return cGCD / 100.0
}

func (p *playerCharacter) applyEffect(eff effect) {
	p.Effects = append(p.Effects, eff)

	globalLog(Basic, fmt.Sprintf("    [BUFF APPLIED] %s to %s: %v %v", eff.Name, p.Name, eff.OffensiveBuffs, eff.DefensiveBuffs))
}

func (p *playerCharacter) reset() {
	p.TotalTime = 0.0
	p.AutoAttackTimeRemaining = 0.0
	p.GCDTimeRemaining = 0.0
	p.IsGCDOnCooldown = false
	p.AnimationLock = 0.0
	p.PlayerStance.reset()
	for i := range p.Skills {
		p.Skills[i].TimeRemainingOnCooldown = 0.0
		p.Skills[i].OnCooldown = false
	}
}

func (p *playerCharacter) getSkillsAvailable() []bool {
	skillAvailable := make([]bool, len(p.Skills))
	for i := range p.Skills {
		if p.Skills[i].IsWeaponskill {
			skillAvailable[i] = !p.IsGCDOnCooldown
		} else {
			skillAvailable[i] = !p.Skills[i].OnCooldown
		}
	}

	return skillAvailable
}

func (p *playerCharacter) isSkillAvailable(name string) bool {
	for i := range p.Skills {
		if p.Skills[i].Name == name {
			// If we are still on GCD cooldown, no Weaponskills allowed
			if p.Skills[i].IsWeaponskill && p.IsGCDOnCooldown {
				return false
			}
			// If the skill is on cooldown, we can't use it
			if p.Skills[i].OnCooldown {
				return false
			}
			// If the skill requires the player's stance to be active and the stance is
			// not active, do nothing
			if p.Skills[i].CheckStance && !p.PlayerStance.checkSkill(name) {
				return false
			}
			// If a skill requires an additional effect to be active and we do not have
			// that effect active, do nothing
			if p.Skills[i].RequiredEffect != "" {
				haveRequiredEffect := false
				for j := range p.Effects {
					if p.Effects[j].Name == p.Skills[i].RequiredEffect {
						haveRequiredEffect = true
						break
					}
				}

				if !haveRequiredEffect {
					return false
				}
			}
			return true
		}
	}
	// Unknown skill name
	return false
}
