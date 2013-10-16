package parser

import (
	. "launchpad.net/gocheck"
	"time"
)

type QueryApiSuite struct{}

var _ = Suite(&QueryApiSuite{})

func (self *QueryApiSuite) TestGetStartTime(c *C) {
	for _, queryStr := range []string{
		"select * from t where time > now() - 1d and time < now() - 1h;",
		"select * from t where now() - 1d < time and time < now() - 1h;",
	} {
		query, err := ParseQuery(queryStr)
		c.Assert(err, IsNil)
		startTime := query.GetStartTime()
		roundedStartTime := startTime.Round(time.Minute).Unix()
		yesterday := time.Now().Add(-24 * time.Hour).Round(time.Minute).Unix()
		c.Assert(roundedStartTime, Equals, yesterday)
		c.Assert(query.GetWhereCondition(), IsNil)
	}
}

func (self *QueryApiSuite) TestGetEndTime(c *C) {
	for _, queryStr := range []string{
		"select * from t where time > now() - 1d and time < now() - 1h;",
		"select * from t where now() - 1d < time and now() - 1h > time;",
	} {
		query, err := ParseQuery(queryStr)
		c.Assert(err, IsNil)
		endTime := query.GetEndTime()
		roundedEndTime := endTime.Round(time.Minute).Unix()
		anHourAgo := time.Now().Add(-1 * time.Hour).Round(time.Minute).Unix()
		c.Assert(roundedEndTime, Equals, anHourAgo)
		c.Assert(query.GetWhereCondition(), IsNil)
	}
}

func (self *QueryApiSuite) TestGetReferencedColumns(c *C) {
	queryStr := "select value1, sum(value2) from t where value > 90.0 and value2 < 10.0 group by value3;"
	query, err := ParseQuery(queryStr)
	c.Assert(err, IsNil)
	columns := query.GetReferencedColumns()
	c.Assert(columns, DeepEquals, map[string][]string{"t": []string{"value", "value1", "value2", "value3"}})
}

func (self *QueryApiSuite) TestGetStartTimeWithOr(c *C) {
	for _, queryStr := range []string{
		"select * from t where time > now() - 1d and (value > 90 or value < 10);",
	} {
		query, err := ParseQuery(queryStr)
		c.Assert(err, IsNil)
		startTime := query.GetStartTime()
		roundedStartTime := startTime.Round(time.Minute).Unix()
		yesterday := time.Now().Add(-24 * time.Hour).Round(time.Minute).Unix()
		c.Assert(roundedStartTime, Equals, yesterday)
	}
}

func (self *QueryApiSuite) TestErrorInStartTime(c *C) {
	queriesAndErrors := map[string]string{
		"select * from t where time > now() * 1d and time < now() - 1h;":  ".*'\\*'.*",
		"select * from t where time > blah * 1d and time < now() - 1h;":   ".*strconv.ParseInt.*",
		"select * from t where time == now() * 1d and time < now() - 1h;": ".*Cannot use time with '=='.*",
		"select * from t where time > now() - 1d or time > now() - 1h;":   ".*Invalid where.*",
		"select * from t where time > foo() - 1d or time > now() - 1h;":   ".*Invalid use of function foo.*",
	}

	for queryStr, error := range queriesAndErrors {
		_, err := ParseQuery(queryStr)
		c.Assert(err, ErrorMatches, error)
	}
}
