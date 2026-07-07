local config_augroup = vim.api.nvim_create_augroup("treesitter_config", { clear = true })

return {
	{
		"nvim-treesitter/nvim-treesitter",
		build = ":TSUpdate",
		config = function()
			local ensure_installed = {
				"c",
				"css",
				"go",
				"gomod",
				"gowork",
				"gosum",
				"html",
				"javascript",
				"lua",
				"markdown",
				"markdown_inline",
				"tsx",
				"typescript",
			}

			local installed = require("nvim-treesitter.config").get_installed("parsers")
			local installed_set = {}
			for _, lang in ipairs(installed) do
				installed_set[lang] = true
			end

			local to_install = {}
			for _, lang in ipairs(ensure_installed) do
				if not installed_set[lang] then
					table.insert(to_install, lang)
				end
			end

			if #to_install > 0 then
				vim.cmd("TSInstall! " .. table.concat(to_install, " "))
			end

			vim.api.nvim_create_autocmd("FileType", {
				pattern = { "*" },
				group = config_augroup,
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
