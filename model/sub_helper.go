package model

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/go-rod/rod/lib/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// OrganizeDlSubFiles 需要从汇总来是网站字幕中，解压对应的压缩包中的字幕出来
func OrganizeDlSubFiles(subInfos []common.SupplierSubInfo) ([]string, error) {

	// 缓存列表，整理后的字幕列表
	var siteSubInfoDict = make([]string, 0)
	tmpFolderFullPath, err := GetTmpFolder()
	if err != nil {
		return nil, err
	}
	// 先清理缓存目录
	err = ClearTmpFolder()
	if err != nil {
		return nil, err
	}
	// 第三方的解压库，首先不支持 io.Reader 的操作，也就是得缓存到本地硬盘再读取解压
	// 且使用 walk 会无法解压 rar，得指定具体的实例，太麻烦了，直接用通用的接口得了，就是得都缓存下来再判断
	// 基于以上两点，写了一堆啰嗦的逻辑···
	for _, subInfo := range subInfos {
		// 先存下来，保存是时候需要前缀，前缀就是从那个网站下载来的
		nowFileSaveFullPath := path.Join(tmpFolderFullPath, GetFrontNameAndOrgName(subInfo))
		err = utils.OutputFile(nowFileSaveFullPath, subInfo.Data)
		if err != nil {
			GetLogger().Errorln("getFrontNameAndOrgName - OutputFile",subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
			continue
		}
		nowExt := strings.ToLower(subInfo.Ext)
		if nowExt != ".zip" && nowExt != ".tar" && nowExt != ".rar" && nowExt != ".7z" {
			// 是否是受支持的字幕类型
			if IsSubExtWanted(nowExt) == false {
				continue
			}
			// 加入缓存列表
			siteSubInfoDict = append(siteSubInfoDict, nowFileSaveFullPath)
		} else {
			// 那么就是需要解压的文件了
			// 解压，给一个单独的文件夹
			unzipTmpFolder := path.Join(tmpFolderFullPath, subInfo.FromWhere)
			err = os.MkdirAll(unzipTmpFolder, os.ModePerm)
			if err != nil {
				return nil, err
			}
			err = UnArchiveFile(nowFileSaveFullPath, unzipTmpFolder)
			// 解压完成后，遍历受支持的字幕列表，加入缓存列表
			if err != nil {
				GetLogger().Errorln("archiver.UnArchive", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := SearchMatchedSubFile(unzipTmpFolder)
			if err != nil {
				GetLogger().Errorln("searchMatchedSubFile", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := AddFrontName(subInfo, filepath.Base(fileFullPath))
				newSubNameFullPath := path.Join(tmpFolderFullPath, newSubName)
				// 改名
				err = os.Rename(fileFullPath, newSubNameFullPath)
				if err != nil {
					GetLogger().Errorln("os.Rename", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
					continue
				}
				// 加入缓存列表
				siteSubInfoDict = append(siteSubInfoDict, newSubNameFullPath)
			}
		}
	}

	return siteSubInfoDict, nil
}