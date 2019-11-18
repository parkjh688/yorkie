package proxy

import (
	time2 "time"

	"github.com/hackerwins/yorkie/pkg/document/change"
	"github.com/hackerwins/yorkie/pkg/document/json"
	"github.com/hackerwins/yorkie/pkg/document/json/datatype"
	"github.com/hackerwins/yorkie/pkg/document/operation"
	"github.com/hackerwins/yorkie/pkg/document/time"
)

type ObjectProxy struct {
	*json.Object
	context *change.Context
}

func ProxyObject(ctx *change.Context, root *json.Object) *ObjectProxy {
	members := datatype.NewRHT()

	for key, val := range root.Members() {
		switch elem := val.(type) {
		case *json.Object:
			members.Set(key, ProxyObject(ctx, elem))
		case *json.Array:
			members.Set(key, ProxyArray(ctx, elem))
		case *datatype.Primitive:
			members.Set(key, elem)
		}
	}

	return NewObjectProxy(ctx, members, root.CreatedAt())
}

func NewObjectProxy(
	ctx *change.Context,
	members *datatype.RHT,
	createdAt *time.Ticket,
) *ObjectProxy {
	return &ObjectProxy{
		Object:  json.NewObject(members, createdAt),
		context: ctx,
	}
}

func (p *ObjectProxy) SetNewObject(k string) *ObjectProxy {
	v := p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return NewObjectProxy(p.context, datatype.NewRHT(), ticket)
	})

	return v.(*ObjectProxy)
}

func (p *ObjectProxy) SetNewArray(k string) *ArrayProxy {
	v := p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return NewArrayProxy(p.context, datatype.NewRGA(), ticket)
	})

	return v.(*ArrayProxy)
}

func (p *ObjectProxy) SetBool(k string, v bool) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) SetInteger(k string, v int) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) SetLong(k string, v int64) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) SetDouble(k string, v float64) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) SetString(k, v string) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) SetBytes(k string, v []byte) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) SetDate(k string, v time2.Time) *ObjectProxy {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return datatype.NewPrimitive(v, ticket)
	})

	return p
}

func (p *ObjectProxy) Remove(k string) datatype.Element {
	removed := p.Object.Remove(k)

	if removed != nil {
		ticket := p.context.IssueTimeTicket()
		p.context.Push(operation.NewRemove(
			p.CreatedAt(),
			removed.CreatedAt(),
			ticket,
		))
	}

	return removed
}

func (p *ObjectProxy) GetArray(k string) *ArrayProxy {
	elem := p.Object.Get(k)
	if elem == nil {
		return nil
	}

	return p.Object.Get(k).(*ArrayProxy)
}

func (p *ObjectProxy) GetObject(k string) *ObjectProxy {
	elem := p.Object.Get(k)
	if elem == nil {
		return nil
	}

	return p.Object.Get(k).(*ObjectProxy)
}

func (p *ObjectProxy) setInternal(
	k string,
	creator func(ticket *time.Ticket) datatype.Element,
) datatype.Element {
	ticket := p.context.IssueTimeTicket()
	value := creator(ticket)
	p.Set(k, value)

	p.context.Push(operation.NewSet(
		p.CreatedAt(),
		k,
		toOriginal(value),
		ticket,
	))

	return value
}