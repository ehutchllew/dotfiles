return {
	{
		"nvim-treesitter/nvim-treesitter",
		build = ":TSUpdate",
		config = function()
			require("nvim-treesitter.config").setup {
				ensure_installed = {
					"c",
					"go",
					"gomod",
					"gowork",
					"gosum",
					"markdown",
					"markdown_inline",
					"query",
					"tsx",
					"typescript",
				},
				auto_install = false,
			}

			vim.api.nvim_create_autocmd("FileType", {
				callback = function(args)
					local ok, stats = pcall(vim.uv.fs_stat, vim.api.nvim_buf_get_name(args.buf))
					if ok and stats and stats.size > 100 * 1024 then
						return
					end
					local lang = vim.treesitter.language.get_lang(vim.bo[args.buf].filetype)
					if lang and pcall(vim.treesitter.language.inspect, lang) then
						vim.treesitter.start(args.buf, lang)
					end
				end,
			})
		end
	},
	{
		"nvim-treesitter/nvim-treesitter-context"
	}
}
