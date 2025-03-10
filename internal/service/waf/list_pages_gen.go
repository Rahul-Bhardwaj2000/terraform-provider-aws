// Code generated by "internal/generate/listpages/main.go -ListOps=ListByteMatchSets,ListGeoMatchSets,ListIPSets,ListRateBasedRules,ListRegexMatchSets,ListRegexPatternSets,ListRuleGroups,ListRules,ListSizeConstraintSets,ListSqlInjectionMatchSets,ListWebACLs,ListXssMatchSets -Paginator=NextMarker -ContextOnly"; DO NOT EDIT.

package waf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
)

func listByteMatchSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListByteMatchSetsInput, fn func(*waf.ListByteMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListByteMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listGeoMatchSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListGeoMatchSetsInput, fn func(*waf.ListGeoMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListGeoMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listIPSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListIPSetsInput, fn func(*waf.ListIPSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListIPSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listRateBasedRulesPages(ctx context.Context, conn *waf.WAF, input *waf.ListRateBasedRulesInput, fn func(*waf.ListRateBasedRulesOutput, bool) bool) error {
	for {
		output, err := conn.ListRateBasedRulesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listRegexMatchSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListRegexMatchSetsInput, fn func(*waf.ListRegexMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListRegexMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listRegexPatternSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListRegexPatternSetsInput, fn func(*waf.ListRegexPatternSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListRegexPatternSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listRuleGroupsPages(ctx context.Context, conn *waf.WAF, input *waf.ListRuleGroupsInput, fn func(*waf.ListRuleGroupsOutput, bool) bool) error {
	for {
		output, err := conn.ListRuleGroupsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listRulesPages(ctx context.Context, conn *waf.WAF, input *waf.ListRulesInput, fn func(*waf.ListRulesOutput, bool) bool) error {
	for {
		output, err := conn.ListRulesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listSizeConstraintSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListSizeConstraintSetsInput, fn func(*waf.ListSizeConstraintSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListSizeConstraintSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listSQLInjectionMatchSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListSqlInjectionMatchSetsInput, fn func(*waf.ListSqlInjectionMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListSqlInjectionMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listWebACLsPages(ctx context.Context, conn *waf.WAF, input *waf.ListWebACLsInput, fn func(*waf.ListWebACLsOutput, bool) bool) error {
	for {
		output, err := conn.ListWebACLsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
func listXSSMatchSetsPages(ctx context.Context, conn *waf.WAF, input *waf.ListXssMatchSetsInput, fn func(*waf.ListXssMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListXssMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
