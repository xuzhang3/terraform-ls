package lang

import (
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-ls/internal/terraform/schema"
	lsp "github.com/sourcegraph/go-lsp"
)

type providerBlockFactory struct {
	logger *log.Logger
	caps   lsp.TextDocumentClientCapabilities

	schemaReader schema.Reader
}

func (f *providerBlockFactory) New(block *hclsyntax.Block) (ConfigBlock, error) {
	if f.logger == nil {
		f.logger = discardLog()
	}

	labels := block.Labels
	if len(labels) != 1 {
		return nil, &invalidLabelsErr{f.BlockType(), labels}
	}

	return &providerBlock{
		hclBlock: block,
		logger:   f.logger,
		caps:     f.caps,
		sr:       f.schemaReader,
	}, nil
}

func (f *providerBlockFactory) BlockType() string {
	return "provider"
}

type providerBlock struct {
	logger   *log.Logger
	caps     lsp.TextDocumentClientCapabilities
	hclBlock *hclsyntax.Block
	sr       schema.Reader
}

func (p *providerBlock) Name() string {
	return p.hclBlock.Labels[0]
}

func (p *providerBlock) BlockType() string {
	return "provider"
}

func (p *providerBlock) CompletionItemsAtPos(pos hcl.Pos) (lsp.CompletionList, error) {
	list := lsp.CompletionList{}

	if p.sr == nil {
		return list, &noSchemaReaderErr{p.BlockType()}
	}

	pSchema, err := p.sr.ProviderConfigSchema(p.Name())
	if err != nil {
		return list, err
	}

	cb := &completableBlock{
		logger:   p.logger,
		caps:     p.caps,
		hclBlock: p.hclBlock,
		schema:   pSchema.Block,
	}
	return cb.completionItemsAtPos(pos)
}
