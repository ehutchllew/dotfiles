local M = {}


local function config()
	return {
		extensions = {
			["ui-select"] = {
				require("telescope.themes").get_dropdown()
			}
		}
	}
end

local function map(keys, func, event, desc)
	vim.keymap.set("n", keys, func, { buffer = event.buf, desc = "LSP: " .. desc })
end

M.setup = function()
	local tsbuiltin = require("telescope.builtin")

	-- Keymaps
	vim.api.nvim_create_autocmd("LspAttach", {
		group = vim.api.nvim_create_augroup("kickstart-lsp-attach", { clear = true }),
		callback = function(event)
			map("<leader>ds", tsbuiltin.lsp_document_symbols, event, "[D]ocument [S]ymbols")
			map("gd", tsbuiltin.lsp_definitions, event, "[G]oto [D]efinition")
			map("gi", tsbuiltin.lsp_implementations, event, "[G]oto [I]mplementation")
			map("gr", tsbuiltin.lsp_references, event, "[G]oto [R]eferences")
			map("gt", tsbuiltin.lsp_type_definitions, event, "[G]oto [T]ype Definition")
			map("K", vim.lsp.buf.hover, event, "Hover Do[K]umentation")
			map("<leader>ws", tsbuiltin.lsp_dynamic_workspace_symbols, event, "[W]ocument [S]ymbols")
		end
	})

	-- Capabilities
	local capabilities = vim.lsp.protocol.make_client_capabilities()
	capabilities = vim.tbl_deep_extend("force", capabilities, require("cmp_nvim_lsp").default_capabilities())

	-- LSP Servers
	-- See `:help lspconfig-all`
	local servers = {
		bashls = {},
		gopls = {},
		jsonls = {},
		lua_ls = {
			settings = {
				Lua = {
					runtime = {
						version = "LuaJIT",
						special = {
							reload = "require"
						}
					},
					workspace = {
						checkThirdParty = false,
						library = {
							require("neodev.config").types(),
							"${3rd}/luv/library",
							unpack(vim.api.nvim_get_runtime_file("", true))
						}
					},
					completion = {
						callSnippet = "Replace"
					}
				}
			}
		},
		tsserver = {},
	}

	-- Ensure servers and tools above are installed
	require("mason").setup()
	require("mason-lspconfig").setup({
		handlers = {
			function(server_name)
				local server = servers[server_name] or {}
				server.capabilities = vim.tbl_deep_extend("force", {}, capabilities, server.capabilities or {})
				require("lspconfig")[server_name].setup(server)
			end
		}
	})
end


return M
