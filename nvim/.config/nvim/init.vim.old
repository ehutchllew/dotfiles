syntax on

set noerrorbells
set tabstop=4 softtabstop=4
set expandtab
set nu rnu
set shiftwidth=4
set smartindent
set noswapfile
set nobackup
set undodir=~/.vim/undodir
set undofile
set incsearch
set cursorcolumn
set cursorline
set lcs+=space:·
set colorcolumn=80
" set run
highlight ColorColumn ctermbg=0 guibg=lightgrey

call plug#begin('~/.vim/plugged')

Plug 'neoclide/coc.nvim', {'branch': 'release'}

Plug 'morhetz/gruvbox'
Plug 'junegunn/seoul256.vim'
Plug 'Rigellute/rigel'

Plug 'preservim/nerdtree'
Plug 'junegunn/fzf', { 'do': { -> fzf#install() } }
Plug 'junegunn/fzf.vim'
Plug 'leafgarland/typescript-vim'
Plug 'pangloss/vim-javascript'
Plug 'vim-airline/vim-airline'
Plug 'vim-airline/vim-airline-themes'

Plug 'ryanoasis/vim-devicons'

call plug#end()

colorscheme gruvbox
set background=dark
set backspace=start,eol,indent

let mapleader = " "

let g:fzf_layout = { 'window': { 'width': 0.8, 'height': 0.8 } }
let $FZF_DEFAULT_OPTS='--reverse'

let g:coc_disable_startup_warning = 1

let g:airline_theme='deus'

let g:webdevicons_enable = 1
let g:webdevicon_enable_nerdtree = 1

nnoremap <leader>b :NERDTreeToggle<CR>
nnoremap <leader>f :NERDTreeFind<CR>

" GOTO remaps
nmap <leader>gd <Plug>(coc-definition)
nmap <leader>gt <Plug>(coc-type-definition)
nmap <leader>gi <Plug>(coc-implementation)
nmap <leader>gr <Plug>(coc-references)

" Window Hopping
nnoremap <leader>h :wincmd h<CR>
nnoremap <leader>j :wincmd j<CR>
nnoremap <leader>k :wincmd k<CR>
nnoremap <leader>l :wincmd l<CR>

" Line Manipulation
nnoremap <A-j> :m+<CR>==
nnoremap <A-k> :m-2<CR>==

" Partition New Buffer
nnoremap <leader>pv :wincmd v<CR>
nnoremap <leader>vs :vs<CR>

" Searching
nnoremap <leader>prw :CocSearch <C-R>=expand("<cword>")<CR><CR>
nnoremap <A-p> :GFiles<CR>
nnoremap * :keepjumps normal! mi*`i<CR>

" Use tab for autocompletion
function! s:check_back_space() abort
    let col = col('.') - 1
    return !col || getline('.')[col - 1] =~ '\s'
endfunction

inoremap <silent><expr> <Tab>
    \ coc#pum#visible() ? "\<C-n>":
    \ <SID>check_back_space() ? "\<Tab>" :
    \ coc#refresh()

inoremap <expr> <Tab> coc#pum#visible() ? "\<C-n>" : "\<Tab>"
inoremap <expr> <S-Tab> coc#pum#visible() ? "\<C-p>" : "\<S-Tab>"
inoremap <expr> <cr> coc#pum#visible() ? "\<C-y>" : "\<C-g>u\<CR>"
