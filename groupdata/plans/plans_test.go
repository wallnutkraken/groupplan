package plans_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wallnutkraken/groupplan/groupdata/plans"
)

func TestPlan_HasTitle_ValidateSucceds(t *testing.T) {
	that := assert.New(t)
	plan := plans.Plan{
		Title:      "ads",
		Identifier: "asd",
	}

	that.NoError(plan.Validate(), "Validate returned an error when it should not have")
}

func TestPlan_TitleWhitespace_ValidateFails(t *testing.T) {
	that := assert.New(t)
	plan := plans.Plan{
		Title:      "        ",
		Identifier: "asd",
	}

	that.Error(plan.Validate(), "Validate returned no error when it should have")
}
