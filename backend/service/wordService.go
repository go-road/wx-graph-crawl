package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/types"
)

// WordService 处理Word文档生成的服务
type WordService struct {
	SavePath string // Word文档保存路径
}

// NewWordService 创建一个新的Word文档生成服务
func NewWordService(savePath string) *WordService {
	return &WordService{
		SavePath: savePath,
	}
}

// GenerateWordForEachArticle 为每篇文章生成一个Word文档
func (svc *WordService) GenerateWordForEachArticle(results []types.CrawlResult) error {
	for _, result := range results {
		if result.WriteContent == "" {
			continue
		}

		// 提取文章标题作为文件名
		//title := extractTitleFromContent(result.WriteContent)
		title := result.Title
		if title == "" {
			title = fmt.Sprintf("文章_%d", result.Number)
		}

		// 处理文件名中的非法字符
		title = sanitizeFilename(title)

		// 构建Word文档路径
		wordFilePath := filepath.Join(svc.SavePath, fmt.Sprintf("%s.docx", title))

		// 生成Word文档
		if err := svc.createWordDocument(result.WriteContent, wordFilePath); err != nil {
			return errors.Wrap(err, fmt.Sprintf("生成Word文档失败: %s", title))
		}
	}
	return nil
}

// 从内容中提取标题
func extractTitleFromContent(content string) string {
	lines := strings.Split(content, "\r\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "标题：") {
			return strings.TrimPrefix(line, "标题：")
		}
	}
	return ""
}

// 处理文件名中的非法字符
func sanitizeFilename(filename string) string {
	// 替换Windows文件名中的非法字符
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, ":", "-")
	filename = strings.ReplaceAll(filename, "\"", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "|", "")
	filename = strings.ReplaceAll(filename, "?", "")
	filename = strings.ReplaceAll(filename, "*", "")

	// 限制文件名长度
	if len(filename) > 100 {
		filename = filename[:100]
	}

	return filename
}

// 创建Word文档
func (svc *WordService) createWordDocument(content, filePath string) error {
	// 确保保存目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "创建保存目录失败")
	}

	// 创建Word文档文件
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "创建Word文档文件失败")
	}
	defer file.Close()

	// 创建ZIP写入器（DOCX是ZIP格式）
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 创建[Content_Types].xml
	if err := svc.createContentTypes(zipWriter); err != nil {
		return err
	}

	// 创建_rels/.rels
	if err := svc.createRelsRels(zipWriter); err != nil {
		return err
	}

	// 创建word/_rels/document.xml.rels
	if err := svc.createDocumentRels(zipWriter); err != nil {
		return err
	}

	// 创建word/document.xml
	if err := svc.createDocumentXml(zipWriter, content); err != nil {
		return err
	}

	// 创建word/styles.xml
	if err := svc.createStylesXml(zipWriter); err != nil {
		return err
	}

	return nil
}

// 创建[Content_Types].xml文件
func (svc *WordService) createContentTypes(zipWriter *zip.Writer) error {
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

	return svc.createZipFile(zipWriter, "[Content_Types].xml", contentTypes)
}

// 创建_rels/.rels文件
func (svc *WordService) createRelsRels(zipWriter *zip.Writer) error {
	relsRels := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	return svc.createZipFile(zipWriter, "_rels/.rels", relsRels)
}

// 创建word/_rels/document.xml.rels文件
func (svc *WordService) createDocumentRels(zipWriter *zip.Writer) error {
	documentRels := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

	return svc.createZipFile(zipWriter, "word/_rels/document.xml.rels", documentRels)
}

// 创建word/document.xml文件
func (svc *WordService) createDocumentXml(zipWriter *zip.Writer, content string) error {
	// 处理内容，将换行符转换为Word中的段落
	paragraphs := strings.Split(content, "\r\n")
	var docContent string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para != "" {
			docContent += fmt.Sprintf("<w:p><w:r><w:t>%s</w:t></w:r></w:p>", escapeXML(para))
		}
	}

	documentXml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    %s
    <w:p/>
  </w:body>
</w:document>`, docContent)

	return svc.createZipFile(zipWriter, "word/document.xml", documentXml)
}

// 创建word/styles.xml文件
func (svc *WordService) createStylesXml(zipWriter *zip.Writer) error {
	stylesXml := `<?xml version="1.0" encoding="UTF-8"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Calibri" w:eastAsia="微软雅黑" w:hAnsi="Calibri" w:cs="Calibri"/>
        <w:sz w:val="22"/>
        <w:szCs w:val="22"/>
      </w:rPr>
    </w:rPrDefault>
  </w:docDefaults>
</w:styles>`

	return svc.createZipFile(zipWriter, "word/styles.xml", stylesXml)
}

// 创建ZIP文件
func (svc *WordService) createZipFile(zipWriter *zip.Writer, filename, content string) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("创建ZIP文件条目失败: %s", filename))
	}

	_, err = io.WriteString(writer, content)
	return errors.Wrap(err, fmt.Sprintf("写入ZIP文件条目失败: %s", filename))
}

// 转义XML特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
