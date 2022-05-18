# 如何协作

## 大纲

以 Github 为例, 假设用户名为 `username`.
其他任何的远程代码仓库同理.

1. Git 项目同步到本地
2. 设置本项目(`https://github.com/lanthora/uranus.git`) 为 upstream
3. 设置自己项目(`https://github.com/username/uranus.git`)为 origin
4. upstream 代码同步到本地 master 分支,同时本地代码同步到 origin
5. 在 master 分支 checkout 功能分支
6. 修改代码,功能分支 push 到自己的 origin.
7. 在 Github 发起功能分支到本项目 master 分支的 PR(Pull Request)
8. 如果存在需要优化的代码,直接修改后 push 到 origin, 无需再次 PR

如果看完上面的流程感觉说的都是废话,说明你可以直接动手PR了,这篇文章是给不熟悉 Git 流程的人准备的

## 代码同步到本地

大纲中的步骤 1 2 3.

如果你使用 Github,可以先 fork, 然后把 fork 到自己账户下的项目 clone 到本地.
此时默认 origin 地址就是自己账号下的项目地址.就完成了大纲中的步骤 1 和 3.

此时只需要添加一个名为 upstream 的 remote 即可.大纲中的步骤 2.

```bash
git remote add upstream https://github.com/lanthora/uranus.git
```

如果你没有在自己账号下 clone, 或者不是 Github 用户.
那么就可以在自己使用的 Git 远程仓库上创建名为 `uranus` 的仓库.
创建仓库后就可以获得git链接.此时依旧假设你的用户名为`username`,
那么 SSH 链接应该是 `git@github.com:username/uranus.git`.
注意 Github 已经不支持通过 HTTPS 上传代码了.

```bash
# 步骤 1
git clone https://github.com/lanthora/uranus.git
# 步骤 2
git remote add upstream https://github.com/lanthora/uranus.git
# 步骤 3
git remote set-url origin git@github.com:username/uranus.git
```

## 与上游保持同步

大纲中的步骤 4.

```bash
git pull upstream master
git push origin master
```

一定不要直接在 master 分支修改代码,否则上面的操作处出现冲突

## 代码推送到远程仓库

大纲中的步骤 5 6 和 8.

创建功能分支并推送到自己的远程仓库.

```bash
git checkout -b feat
git push -u origin feat
```

对 commit, push 不做赘述.

## 发起 PR

大纲中的步骤 7.

可以看出步骤 6 7 8 没有严格的顺序关系.
有些人喜欢一切准备就绪再发起 PR.
有些人喜欢先发起空 PR 再逐步实现功能.
在任何你认为合适的时间发起 PR 即可.
