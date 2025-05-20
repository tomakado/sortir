package checker

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.tomakado.io/sortir/internal/config"
	"golang.org/x/tools/go/analysis"
)

type mockNode struct{}

func (m *mockNode) Pos() token.Pos { return token.NoPos }
func (m *mockNode) End() token.Pos { return token.NoPos }

type CheckerTestSuite struct {
	suite.Suite
}

func (s *CheckerTestSuite) testElementsSorted(groups [][]metadata, prefix, globalPrefix, message string, shouldPass bool) []analysis.Diagnostic {
	s.T().Parallel()

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, prefix, message)
	s.Require().Equal(shouldPass, result)
	return reported
}

func (s *CheckerTestSuite) TestCheckElementsSorted_AllSortedSingleGroup() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "a", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "b", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "c", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_NotSortedSingleGroup() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "b", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "c", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().False(result)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(2), reported[0].Pos)
	s.Require().Equal("test message", reported[0].Message)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_MultipleGroups() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "a", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "b", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
		},
		{
			{Value: "d", Position: token.Pos(4), Line: 4, Node: &mockNode{}},
			{Value: "c", Position: token.Pos(5), Line: 5, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().False(result)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(5), reported[0].Pos)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_EmptyGroups() {
	s.T().Parallel()

	groups := [][]metadata{{}}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_SingleElementGroups() {
	s.T().Parallel()

	groups := [][]metadata{
		{{Value: "c", Position: token.Pos(1), Line: 1, Node: &mockNode{}}},
		{{Value: "a", Position: token.Pos(2), Line: 2, Node: &mockNode{}}},
		{{Value: "b", Position: token.Pos(3), Line: 3, Node: &mockNode{}}},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_WithPrefixFiltering() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "prefixA", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "notPrefixed", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "prefixB", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "prefix", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_WithPrefixFilteringNotSorted() {
	groups := [][]metadata{
		{
			{Value: "a", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "prefixB", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "z", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
			{Value: "prefixA", Position: token.Pos(4), Line: 4, Node: &mockNode{}},
		},
	}

	reported := s.testElementsSorted(groups, "prefix", "", "test message", false)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(4), reported[0].Pos)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_WithGlobalPrefix() {
	groups := [][]metadata{
		{
			{Value: "globalPrefixB", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "otherThing", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "globalPrefixA", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := s.testElementsSorted(groups, "", "globalPrefix", "test message", false)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(3), reported[0].Pos)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_PrefixOverridesGlobal() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "localPrefixA", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "globalPrefixZ", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "localPrefixB", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "prefix", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_EmptyStringValues() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "b", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_AllElementsFilteredOut() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "otherA", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "otherB", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "otherC", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_MixedPrefixedAndNonPrefixed() {
	groups := [][]metadata{
		{
			{Value: "otherA", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "prefixB", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "z", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
			{Value: "prefixA", Position: token.Pos(4), Line: 4, Node: &mockNode{}},
		},
	}

	reported := s.testElementsSorted(groups, "prefix", "", "test message", false)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(4), reported[0].Pos)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_MultipleSortingIssuesReportsFirst() {
	groups := [][]metadata{
		{
			{Value: "c", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "b", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := s.testElementsSorted(groups, "", "", "test message", false)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(2), reported[0].Pos)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_NilGroups() {
	s.T().Parallel()

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, nil, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_EmptyPrefixEmptyGlobalPrefix() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "b", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().False(result)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(2), reported[0].Pos)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_IdenticalValues() {
	s.T().Parallel()

	groups := [][]metadata{
		{
			{Value: "a", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	c := New(config.New())

	result := c.checkElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *CheckerTestSuite) TestCheckElementsSorted_BreakAfterFirstReport() {
	groups := [][]metadata{
		{
			{Value: "c", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "b", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := s.testElementsSorted(groups, "", "", "test message", false)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(2), reported[0].Pos)
}

func TestCheckerTestSuite(t *testing.T) {
	suite.Run(t, new(CheckerTestSuite))
}
