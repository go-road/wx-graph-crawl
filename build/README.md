# 构建目录

构建目录用于存放应用程序的所有构建文件和资源。

目录结构如下：

* bin - 输出目录
* darwin - macOS 特定文件
* windows - Windows 特定文件

## Mac

`darwin` 目录包含特定于 Mac 构建的文件。
这些文件可以自定义并作为构建的一部分使用。要将这些文件恢复为默认状态，只需删除它们并使用 `wails build` 进行构建。

该目录包含以下文件：

- `Info.plist` - Mac 构建使用的主 plist 文件。它在使用 `wails build` 构建时被使用。
- `Info.dev.plist` - 与主 plist 文件相同，但在使用 `wails dev` 构建时使用。

## Windows

`windows` 目录包含使用 `wails build` 构建时所需的 manifest 和 rc 文件。
这些文件可以为您的应用程序自定义。要将这些文件恢复为默认状态，只需删除它们并使用 `wails build` 进行构建。

- `icon.ico` - 应用程序使用的图标文件。在使用 `wails build` 构建时会使用此文件。如果您希望使用其他图标，只需用您自己的文件替换此文件。如果文件丢失，将使用构建目录中的 `appicon.png` 文件创建一个新的 `icon.ico` 文件。
- `installer/*` - 用于创建 Windows 安装程序的文件。这些文件在使用 `wails build` 构建时被使用。
- `info.json` - Windows 构建使用的应用程序详细信息。此处的数据将被 Windows 安装程序以及应用程序本身使用（右键单击 exe -> 属性 -> 详细信息）。
- `wails.exe.manifest` - 主应用程序的 manifest 文件。