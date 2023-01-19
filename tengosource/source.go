package tengosource

import (
	"sync"

	"github.com/d5/tengo/v2"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/tengolib/tengodb"
)

const (
	PROVIDER_SQL_MEMORY = "SQL_MEMORY"
	PROVIDER_SQL        = "SQL"
	PROVIDER_CURL       = "CURL"
	PROVIDER_BIN        = "BIN"
	PROVIDER_REDIS      = "REDIS"
	PROVIDER_RABBITMQ   = "RABBITMQ"
)

//IdentiferRelation 维护映射关系，比如将模板名称映射资源标识，实现模板名称找到资源
type IdentiferRelation struct {
	TemplateName    string `json:"templateName"`
	SourceIdentifer string `json:"sourceIdentifer"`
}
type IdentiferRelationCollection []IdentiferRelation

//Add 确保templateName 唯一
func (rc *IdentiferRelationCollection) AddTemplateIdentiferRelation(templateIdentifer string, sourceIdentifer string) (err error) {
	for _, r := range *rc {
		if r.TemplateName != "" && r.TemplateName == templateIdentifer {
			err = errors.Errorf("template identifer:%s exists", templateIdentifer)
			return err
		}
	}
	r := IdentiferRelation{
		TemplateName:    templateIdentifer,
		SourceIdentifer: sourceIdentifer,
	}
	*rc = append(*rc, r)
	return nil
}
func (rc *IdentiferRelationCollection) GetSourceIdentiferByTemplateIdentifer(templateIdentifer string) (sourceIdentifer string, err error) {
	for _, r := range *rc {
		if r.TemplateName != "" && r.TemplateName == templateIdentifer {
			return r.SourceIdentifer, nil
		}
	}
	err = errors.Errorf("not found source identifer by template identifer: %s", templateIdentifer)
	return "", err

}

type SourcePool struct {
	sourceMap                   map[string]Source
	IdentiferRelationCollection IdentiferRelationCollection
	lock                        sync.Mutex
}

//NewSourcePool 生成资源池
func NewSourcePool() (p *SourcePool) {
	p = &SourcePool{
		sourceMap:                   make(map[string]Source),
		IdentiferRelationCollection: make(IdentiferRelationCollection, 0),
	}
	return p
}

type Source struct {
	Identifer string
	Type      string
	Config    string
	provider  tengo.Object
}

//SetProvider 方便外部替换修改(如替换成内存实现提供者)
func (s *Source) SetProvider(provider tengo.Object) {
	s.provider = provider
}

//MakeSource 创建常规资源,方便外部统一调用
func MakeSource(identifer string, typ string, config string) (s Source, err error) {
	s = Source{
		Identifer: identifer,
		Type:      typ,
		Config:    config,
	}
	var provider tengo.Object
	switch s.Type {
	case PROVIDER_SQL:
		provider, err = tengodb.NewTengoDB(s.Config)
		if err != nil {
			return s, err
		}
		//todo curl , bin 提供者实现
	}
	s.provider = provider
	return s, nil
}

func (p *SourcePool) RegisterSource(s Source) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.sourceMap[s.Identifer] = s
	return nil
}

func (p *SourcePool) AddTemplateIdentiferRelation(templateIdentifer string, sourceIdentifer string) (err error) {
	// 资源必须先注册
	_, ok := p.sourceMap[sourceIdentifer]
	if !ok {
		err = errors.Errorf("register source(%s) befor AddTemplateIdentiferRelation", sourceIdentifer)
		return err
	}
	err = p.IdentiferRelationCollection.AddTemplateIdentiferRelation(templateIdentifer, sourceIdentifer)
	if err != nil {
		return err
	}
	return nil
}

func (p *SourcePool) GetProviderBySourceIdentifer(sourceIdentifer string) (sourceProvider tengo.Object, err error) {
	source, ok := p.sourceMap[sourceIdentifer]
	if !ok {
		err = errors.Errorf("not found source by source identifier: %s", sourceIdentifer)
		return nil, err
	}
	sourceProvider = source.provider
	return sourceProvider, nil
}
func (p *SourcePool) GetProviderByTemplateIdentifer(templateIdentifier string) (sourceProvider tengo.Object, err error) {
	sourceIdentifer, err := p.IdentiferRelationCollection.GetSourceIdentiferByTemplateIdentifer(templateIdentifier)
	if err != nil {
		return nil, err
	}
	source, ok := p.sourceMap[sourceIdentifer]
	if !ok {
		err = errors.Errorf("not found source by template identifier: %s", templateIdentifier)
		return nil, err
	}
	sourceProvider = source.provider
	return sourceProvider, nil
}

//TengoGetProviderBySourceIdentifer 注入到tengo 脚本，用来获取资源执行器
func (p *SourcePool) TengoGetProviderBySourceIdentifer(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	sourceIdentiferObj := args[0]
	sourceIdentifer, ok := tengo.ToString(sourceIdentiferObj)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "sourceIdentifer",
			Expected: "string",
			Found:    sourceIdentiferObj.TypeName(),
		}
	}
	sourceProvider, err := p.GetProviderBySourceIdentifer(sourceIdentifer)
	if err != nil {
		return nil, err
	}
	return sourceProvider, nil
}

//TengoGetProviderByTemplateIdentifer 注入到tengo 脚本，用来通过目标标识获取资源执行器
func (p *SourcePool) TengoGetProviderByTemplateIdentifer(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	templateIdentiferObj := args[0]
	templateIdentifer, ok := tengo.ToString(templateIdentiferObj)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "templateIdentifer",
			Expected: "string",
			Found:    templateIdentiferObj.TypeName(),
		}
	}
	sourceProvider, err := p.GetProviderByTemplateIdentifer(templateIdentifer)
	if err != nil {
		return nil, err
	}
	return sourceProvider, nil
}
