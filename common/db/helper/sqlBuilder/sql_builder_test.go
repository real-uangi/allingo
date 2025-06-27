/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/7/31 12:41
 */

// Package sqlBuilder
package sqlBuilder

import "testing"

func TestSqlBuilder(t *testing.T) {
	sb := New()
	sb.Where("")
	sb.Where("is_del = 0")
	sb.Where("id = 1")
	sb.Wheref("text like '%%%d%%'", 123)
	sb.From("")
	sb.From("a_base a")
	sb.Fromf("left join b_base b on a_base.id = b_base.id and a.org_id = '%s'", "qwe")
	sb.Order("id desc", "created_time asc")
	sb.Field("", "id", "order")
	sb.FieldAs("id", "myId")
	sb.Limit(10)
	sb.Offset(20)
	t.Log(sb.StringForCount())
	t.Log(sb.String())
}

func TestNoOrder(t *testing.T) {
	builder := New()
	builder.Field("count(1)")
	builder.From("tarkov_item")
	builder.Where("is_del = 0")
	builder.Where("mode = 1")
	t.Log(builder.String())
}

func TestAdvanced(t *testing.T) {
	builder := New()
	builder.Field(`t1.id as "id"`, `t1.name as "name"`, `sum(t2.id) as "articles"`, `sum(t2.views) as "heat"`)
	builder.From("cms_categories t1").From("left join cms_articles t2 on t2.category_id = t1.id")
	builder.Where("t1.deleted_at IS NULL").Where("t2.deleted_at IS NULL")
	builder.Limit(10)
	builder.GroupBy("t1.id", "t1.name")
	t.Log(builder.String())
	t.Log(builder.StringForCount())
}
