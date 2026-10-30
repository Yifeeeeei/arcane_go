package card_maker

import (
	"encoding/json"
	"log"
	"reflect"
)

// CardInfo is the canonical schema used by the redesigned base set. Optional
// numeric cells are pointers so blank and an explicitly printed zero differ.
type CardInfo struct {
	Number          string   `json:"number"`
	SetCode         string   `json:"set_code"`
	SetName         string   `json:"set_name"`
	Type            string   `json:"type"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Subtype         string   `json:"subtype"`
	Tag             string   `json:"-"`
	Rarity          string   `json:"rarity,omitempty"`
	ManaCost        *int     `json:"mana_cost,omitempty"`
	ActivationMana  *int     `json:"activation_mana,omitempty"`
	TechRequirement Elements `json:"tech_requirement,omitempty"`
	ElementsGain    Elements `json:"elements_gain,omitempty"`
	AttackValue     *int     `json:"attack,omitempty"`
	LifeValue       *int     `json:"life,omitempty"`
	Cooldown        *int     `json:"cooldown,omitempty"`
	SkillSlots      *int     `json:"skill_slots,omitempty"`
	Description     string   `json:"description"`
	Quote           string   `json:"quote"`
	Source          string   `json:"source,omitempty"`
	ArtRef          string   `json:"art_ref,omitempty"`
	OutputPath      string   `json:"output_path"`

	// Deprecated renderer adapter. New imports populate these from the fields
	// above; they are deliberately excluded from the exported schema.
	ElementsCost    Elements `json:"-"`
	ElementsExpense Elements `json:"-"`
	Attack          int      `json:"-"`
	Life            int      `json:"-"`
	Duration        int      `json:"-"`
	Power           int      `json:"-"`
	Spawns          []string `json:"-"`
	VersionNumber   string   `json:"-"`
	VersionName     string   `json:"-"`
	Legacy          bool     `json:"-"`
}

func (cardInfo CardInfo) Equals(other CardInfo) bool { return reflect.DeepEqual(cardInfo, other) }

func (cardInfo CardInfo) String() string {
	b, err := json.MarshalIndent(cardInfo, "", "  ")
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(b)
}
