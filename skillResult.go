package main

import (
	"fmt"
	"math"
)

type skillResult struct {
	SkillName      string
	SkillIndex     int
	DamageDone     int
	BaseDamageDone float64
	DidCrit        bool
	DidDirect      bool
	IsCombo        bool
	CritRate       float64
	CritMultiplier float64
	DirectRate     float64
	DamageBuff     float64
	EffectApplied  []string
	TimePerformed  float64
	PlayerName     string
	TargetName     string
}

func (s *skillResult) Log() {
	modifiers := ""
	if s.IsCombo {
		modifiers += "[COMBO]"
	}
	if s.DidCrit {
		modifiers += "[CRIT]"
	}
	if s.DidDirect {
		modifiers += "[DIRECT]"
	}
	globalLog(Basic, fmt.Sprintf("[%06.2f] [%s -> %s] %s : %d %s", s.TimePerformed, s.PlayerName, s.TargetName, s.SkillName, s.DamageDone, modifiers))
	if s.DamageDone != 0 {
		globalLog(Basic, fmt.Sprintf("    [Base damage:%d] [Attack Buff:%.2fx] [Rates:%.2f,%.2f]", int(s.BaseDamageDone), s.DamageBuff, s.CritRate, s.DirectRate))
	}
}

func (s *skillResult) ApplyBuffs(playerData playerCharacter) {
	for i := range s.EffectApplied {
		applyAdditionalEffect(s.EffectApplied[i], playerData, s.TargetName)
	}
}

func (s *skillResult) calculateDoTTick(dot dotEffect) {
	critRng := rng.Float64() * 100.0
	directRng := rng.Float64() * 100.0

	s.BaseDamageDone = float64(dot.BaseDamage)
	s.DamageDone = dot.BaseDamage
	s.CritRate = dot.CritChance
	s.DirectRate = dot.DirectChance

	if dot.CritChance >= critRng {
		s.DidCrit = true
		s.DamageDone = dot.CritDamage
	}

	if dot.DirectChance >= directRng {
		s.DidDirect = true
		s.DamageDone = int(math.Floor(float64(s.DamageDone) * 1.20))
	}
}

func (s *skillResult) calculateDamage(skill *playerSkill, playerData playerCharacter, isAutoAttack bool, dotPotency int) {
	isWeaponSkill := false
	if isAutoAttack {
		s.BaseDamageDone = autoAttackPot
	} else if dotPotency != 0 {
		s.BaseDamageDone = float64(dotPotency) / 100.0
	} else {
		s.calculatePotency(skill, playerData.LastWeaponskill, playerData.PlayerStance.checkPotencyBoost(skill.Name))
		isWeaponSkill = skill.IsWeaponskill
	}
	s.appendWeaponDamage(isAutoAttack, playerData.WeaponDamage, playerData.WeaponDelay)
	s.appendAttackPower(playerData.AttackPower)
	s.appendDetermination(playerData.Determination)
	s.floorBaseDamage()
	if isAutoAttack || dotPotency != 0 {
		s.appendSkillSpeed(playerData.SkillSpeed)
		s.floorBaseDamage()
	}

	critMod := 0.0
	critMod, s.CritMultiplier = s.calculateCrit(isWeaponSkill, playerData)
	directMod := s.calculateDirect(playerData.DirectHit)
	s.calculateAttackBuff(playerData.Effects)
	s.DamageDone = int(math.Floor(s.BaseDamageDone * critMod))
	s.DamageDone = int(math.Floor(float64(s.DamageDone) * directMod))
	// Slight bug, Floor should be applied after each buff
	s.DamageDone = int(math.Floor(float64(s.DamageDone) * s.DamageBuff))
}

// Returns (damage modifier, isCombo)
func (s *skillResult) calculatePotency(skill *playerSkill, lastWeaponSkill string, potencyBoost float64) {
	isCombo := false
	pot := skill.Potency
	if skill.IsWeaponskill && skill.ComboAction != "" && lastWeaponSkill == skill.ComboAction {
		isCombo = true
		pot = skill.ComboPotency
	}

	s.BaseDamageDone = float64(pot) / 100.0 * potencyBoost
	s.IsCombo = isCombo
}

func (s *skillResult) appendWeaponDamage(autoAttack bool, weaponDamage int, weaponDelay float64) {
	wd := math.Floor((float64(level70Main) * float64(level70DrgStr) / 1000.0) + float64(weaponDamage))
	if autoAttack {
		wd = math.Floor(wd * (weaponDelay / 3))
	}

	s.BaseDamageDone *= wd
}

func (s *skillResult) appendAttackPower(attackPower int) {
	ap := math.Floor((125*(float64(attackPower)-292)/292)+100.0) / 100.0
	s.BaseDamageDone *= ap
}

func (s *skillResult) appendDetermination(determination int) {
	det := math.Floor(130.0*(float64(determination)-float64(level70Main))/float64(level70Div)+1000) / 1000.0
	s.BaseDamageDone *= det
}

func (s *skillResult) appendSkillSpeed(skillSpeed int) {
	ss := math.Floor(130.0*(float64(skillSpeed)-float64(level70Sub))/float64(level70Div)+1000) / 1000.0
	s.BaseDamageDone *= ss
}

// Returns (damage modifier, crit bonus)
func (s *skillResult) calculateCrit(isWeaponSkill bool, playerData playerCharacter) (float64, float64) {
	critBonus := s.calculateCritRate(playerData.CriticalHit)

	s.calculateCritRateBuff(isWeaponSkill, playerData.Effects)

	critRng := rng.Float64() * 100.0
	if s.CritRate >= critRng {
		s.DidCrit = true
		return critBonus, critBonus
	}

	s.DidCrit = false
	return 1.0, critBonus
}

// Crit Rate buffs are additive
func (s *skillResult) calculateCritRateBuff(isWeaponSkill bool, effects []effect) {
	totalBuff := 0.0
	for i := range effects {
		if len(effects[i].OffensiveBuffs) > CritRate {
			if effects[i].AppliesTo == All || (effects[i].AppliesTo == Weaponskill && isWeaponSkill) {
				totalBuff += float64(effects[i].OffensiveBuffs[CritRate])
			}
		}
	}

	s.CritRate += totalBuff
}

// Returns (crit bonus)
func (s *skillResult) calculateCritRate(criticalHit int) float64 {
	s.CritRate = math.Floor(200*((float64(criticalHit)-float64(level70Sub))/float64(level70Div))+50.0) / 10.0
	critBonus := math.Floor(200*(float64(criticalHit)-float64(level70Sub))/float64(level70Div)+1400) / 1000.0
	return critBonus
}

// Returns (damage modifier)
func (s *skillResult) calculateDirect(directHit int) float64 {
	directRng := rng.Float64() * 100.0
	s.calculateDirectRate(directHit)

	if s.DirectRate >= directRng {
		s.DidDirect = true
		return 1.2
	}
	s.DidDirect = false
	return 1.0
}

func (s *skillResult) calculateDirectRate(directHit int) {
	s.DirectRate = math.Floor(550*((float64(directHit)-float64(level70Sub))/float64(level70Div))) / 10.0
}

// Attack buffs are multiplicative
func (s *skillResult) calculateAttackBuff(effects []effect) {
	totalBuff := 1.0
	for i := range effects {
		if len(effects[i].OffensiveBuffs) > DamageDealt {
			buff := 100.0 + float64(effects[i].OffensiveBuffs[DamageDealt])
			buff = buff / 100
			totalBuff = totalBuff * buff
		}
	}

	s.DamageBuff = totalBuff
}

func (s *skillResult) floorBaseDamage() {
	s.BaseDamageDone = math.Floor(s.BaseDamageDone)
}
