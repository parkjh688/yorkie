package proxy

import (
	"github.com/hackerwins/rottie/pkg/document/change"
	"github.com/hackerwins/rottie/pkg/document/json"
	"github.com/hackerwins/rottie/pkg/document/json/datatype"
	"github.com/hackerwins/rottie/pkg/document/operation"
	"github.com/hackerwins/rottie/pkg/document/time"
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
		case *json.Primitive:
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

func (p *ObjectProxy) SetString(k, v string) {
	p.setInternal(k, func(ticket *time.Ticket) datatype.Element {
		return json.NewPrimitive(v, ticket)
	})
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

func (p *ObjectProxy) GetArray(k string) *ArrayProxy {
	elem := p.Object.Get(k)
	if elem == nil {
		return nil
	}

	return p.Object.Get(k).(*ArrayProxy)
}

func (p *ObjectProxy) setInternal(
	k string,
	creator func(ticket *time.Ticket) datatype.Element,
) datatype.Element {
	ticket := p.context.IssueTimeTicket()
	value := creator(ticket)
	p.Set(k, value)

	p.context.Push(operation.NewSet(
		k,
		toOriginal(value),
		p.CreatedAt(),
		ticket,
	))

	return value
}

