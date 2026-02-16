const queryInput = document.getElementById("queryInput");
const refreshBtn = document.getElementById("refreshBtn");
const paperList = document.getElementById("paperList");
const categorySelect = document.getElementById("categorySelect");
const sortSelect = document.getElementById("sortSelect");
const totalCount = document.getElementById("totalCount");
const statsText = document.getElementById("statsText");
const template = document.getElementById("paperCardTemplate");

const monthInput = document.getElementById("monthInput");
const topTierCheckbox = document.getElementById("topTierCheckbox");
const ccfSelect = document.getElementById("ccfSelect");

// Modal Elements
const searchModal = document.getElementById("searchModal");
const closeBtn = document.querySelector(".close-btn");

if (closeBtn) {
    closeBtn.onclick = function() {
        searchModal.style.display = "none";
    }
}

window.onclick = function(event) {
    if (event.target == searchModal) {
        searchModal.style.display = "none";
    }
}

// 全局函数需要尽早定义
window.toggleTrend = function(id, btn) {
    console.log("toggleTrend called for", id);
    const el = document.getElementById(id);
    if (!el) {
        console.error("Element not found:", id);
        return;
    }
    
    if (el.classList.contains('collapsed')) {
        el.classList.remove('collapsed');
        btn.textContent = "收起";
    } else {
        el.classList.add('collapsed');
        btn.textContent = "展开全文";
    }
};

window.searchByTopic = function(topic) {
  const queryInput = document.getElementById("queryInput");
  if (queryInput) {
    queryInput.value = topic;
    fetchPapers(false);
    const list = document.getElementById("paperList");
    if (list) {
        const y = list.getBoundingClientRect().top + window.scrollY - 100;
        window.scrollTo({top: y, behavior: 'smooth'});
    }
  }
};

let allPapers = [];
let activeCategory = "";
let activeSort = "published_desc";
let activeCCF = "";
let currentTranslation = "";

// 选择所有类名为 'source-checkbox' 的复选框
const sourceCheckboxes = Array.from(
  document.querySelectorAll(".source-checkbox")
);

function buildQueryParams() {
  const params = new URLSearchParams();
  const queryValue = queryInput.value.trim();
  if (queryValue) {
    params.set("query", queryValue);
  }
  
  const sources = sourceCheckboxes
    .filter((checkbox) => checkbox.checked)
    .map((checkbox) => checkbox.value);
  sources.forEach((source) => params.append("sources", source));
  
  // limit 在 fetchPapers 中设置
  params.set("sort", activeSort);
  
  if (activeCategory) {
    params.set("category", activeCategory);
  }
  
  if (monthInput.value) {
    params.set("month", monthInput.value);
  }
  
  if (topTierCheckbox.checked) {
    params.set("top_tier", "true");
  }
  
  if (ccfSelect.value) {
    params.set("ccf_level", ccfSelect.value);
  }
  
  return params;
}

function formatAuthors(authors) {
  if (!authors || authors.length === 0) {
    return "作者未知";
  }
  if (authors.length <= 3) {
    return authors.join(", ");
  }
  return `${authors.slice(0, 3).join(", ")} 等`;
}

function formatDate(value) {
  if (!value) {
    return "日期未知";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toISOString().slice(0, 10);
}

function sortPapers(items) {
  const sortValue = sortSelect.value;
  if (sortValue === "published_asc") {
    return [...items].sort((a, b) => new Date(a.published_at) - new Date(b.published_at));
  }
  // 引用量排序已移除
  if (sortValue === "title_asc") {
    return [...items].sort((a, b) => a.title.localeCompare(b.title));
  }
  // 默认：按发布时间降序
  return [...items].sort((a, b) => new Date(b.published_at) - new Date(a.published_at));
}

function applyFilters() {
  let items = [...allPapers];
  
  // 客户端类别过滤（如果需要，尽管 API 可能会处理部分）
  const catVal = categorySelect.value;
  if (catVal) {
    items = items.filter((paper) =>
      (paper.categories || []).some((category) =>
        category.toLowerCase().includes(catVal.toLowerCase())
      )
    );
  }
  
  items = sortPapers(items);
  renderPapers(items);
}

function updateCategoryOptions() {
  const categories = new Set();
  allPapers.forEach((paper) => {
    (paper.categories || []).forEach((category) => categories.add(category));
  });
  
  const currentValue = categorySelect.value;
  categorySelect.innerHTML = '<option value="">全部领域</option>';
  
  Array.from(categories)
    .sort((a, b) => a.localeCompare(b))
    .forEach((category) => {
      const option = document.createElement("option");
      option.value = category;
      option.textContent = category;
      if (category === currentValue) {
        option.selected = true;
      }
      categorySelect.appendChild(option);
    });
}

function updateStats(items) {
  totalCount.textContent = `${items.length} 篇论文`;
  
  // 计算副标题的来源分布
  const stats = items.reduce((acc, item) => {
    const source = item.source || "unknown";
    acc[source] = (acc[source] || 0) + 1;
    return acc;
  }, {});
  
  const statsString = Object.entries(stats)
    .map(([source, count]) => `${source}: ${count}`)
    .join(" · ");
    
  let prefix = "";
  if (currentTranslation) {
      prefix = `[翻译: ${currentTranslation}] `;
  }
  
  statsText.textContent = `${prefix}${statsString ? `来源分布: ${statsString}` : "找到相关结果"}`;
}

function renderPapers(items) {
  paperList.innerHTML = "";
  updateStats(items);
  
  if (items.length === 0) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.textContent = "没有匹配的论文，请调整搜索条件";
    paperList.appendChild(empty);
    return;
  }
  
  items.forEach((paper) => {
    const node = template.content.cloneNode(true);
    const card = node.querySelector(".card");
    
    // 头部信息
    card.querySelector(".source-badge").textContent = paper.source || "unknown";
    card.querySelector(".date-text").textContent = formatDate(paper.published_at);
    
    // CCF 标签
    const ccfClass = paper.ccf_class || "None";
    if (ccfClass !== "None") {
        const ccfTag = document.createElement("span");
        ccfTag.className = `ccf-tag ccf-${ccfClass}`;
        ccfTag.textContent = `CCF ${ccfClass}`;
        card.querySelector(".ccf-container").appendChild(ccfTag);
    }

    // 主要内容
    card.querySelector(".card-title").textContent = paper.title || "无标题";
    card.querySelector(".card-meta").textContent = `${formatAuthors(paper.authors)} · ${
      paper.venue || "未知来源"
    }`;
    
    card.querySelector(".card-abstract").textContent =
      paper.abstract || "暂无摘要信息";
      
    // 类别
    const categoriesDiv = card.querySelector(".categories");
    categoriesDiv.innerHTML = "";
    (paper.categories || []).slice(0, 5).forEach((category) => {
      const tag = document.createElement("span");
      tag.className = "tag";
      tag.textContent = category;
      categoriesDiv.appendChild(tag);
    });
    
    // 链接
    const link = card.querySelector(".read-btn");
    link.href = paper.url || "#";
    
    paperList.appendChild(node);
  });

  // 添加无限滚动哨兵
  const sentinel = document.createElement("div");
  sentinel.id = "scroll-sentinel";
  sentinel.style.height = "20px";
  sentinel.style.width = "100%";
  paperList.appendChild(sentinel);
  
  if (window.observer) {
      window.observer.observe(sentinel);
  }
}

let currentOffset = 0;
let isLoading = false;
const LIMIT = 50;

// 无限滚动观察器
window.observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && !isLoading && allPapers.length > 0) {
        fetchPapers(true);
    }
}, { rootMargin: "200px" });

async function fetchPapers(isAppend = false) {
  if (isLoading) return;

  if (!isAppend) {
      if (searchModal) searchModal.style.display = "block"; // Show modal for new searches
      currentOffset = 0;
      allPapers = [];
      paperList.innerHTML = `<div class="loading">正在加载数据...</div>`;
  } else {
      // 不清空列表，可能在底部显示小加载器
      const loader = document.createElement("div");
      loader.className = "loading-more";
      loader.textContent = "正在加载更多...";
      loader.style.textAlign = "center";
      loader.style.padding = "10px";
      loader.style.color = "#666";
      loader.id = "loading-more-indicator";
      paperList.appendChild(loader);
  }
  
  isLoading = true;

  // UI 加载状态
  const originalBtnContent = refreshBtn.innerHTML;
  if (!isAppend) {
      refreshBtn.disabled = true;
      refreshBtn.innerHTML = `...`; 
  }
  
  try {
    const params = buildQueryParams();
    params.set("offset", currentOffset);
    params.set("limit", LIMIT);

    // 更正 API 端点以匹配 main.go
    const response = await fetch(`/search?${params.toString()}`);
    
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    const newItems = data.items || [];
    
    if (!isAppend) {
        currentTranslation = data.translation || "";
    }
    
    if (isAppend) {
         const loader = document.getElementById("loading-more-indicator");
         if (loader) loader.remove();
         
         if (newItems.length === 0) {
             // 结果结束，如果存在则移除哨兵以停止触发
             const sentinel = document.getElementById("scroll-sentinel");
             if (sentinel) sentinel.remove();
             return; 
         }
    }
    
    allPapers = isAppend ? [...allPapers, ...newItems] : newItems;
    currentOffset += LIMIT; // Advance offset
    
    // 基于所有加载的论文更新类别
    updateCategoryOptions();
    
    applyFilters();
  } catch (error) {
    console.error("Fetch error:", error);
    if (!isAppend) {
        paperList.innerHTML = "<div class='empty-state'>加载失败，请检查网络或稍后再试</div>";
    }
  } finally {
    isLoading = false;
    if (!isAppend) {
        refreshBtn.disabled = false;
        refreshBtn.innerHTML = originalBtnContent;
    }
  }
}

// Daily Summary Logic
async function fetchDailySummary() {
  const container = document.getElementById("dailySummarySection");
  if (!container) return;

  try {
    const response = await fetch("/daily-summary");
    if (!response.ok) throw new Error("Summary fetch failed");
    
    const summary = await response.json();
    renderDailySummary(summary);
  } catch (error) {
    console.error(error);
    container.innerHTML = `<div class="empty-state">今日前沿摘要加载失败，请刷新重试</div>`;
  }
}

function renderDailySummary(summary) {
  const container = document.getElementById("dailySummarySection");
  
  // 日期格式化
  const dateStr = summary.date; // YYYY-MM-DD
  
  // 构建 HTML
  let trendsHtml = (summary.major_trends || []).map((t, index) => {
    const isLong = t.length > 180; 
    
    if (isLong) {
      return `<li class="summary-item">
        <div class="summary-content collapsed" id="trend-${index}">${t}</div>
        <span class="expand-btn" onclick="window.toggleTrend('trend-${index}', this)">展开全文</span>
      </li>`;
    } else {
      return `<li class="summary-item">
        <div class="summary-content">${t}</div>
      </li>`;
    }
  }).join("");

  if (!trendsHtml) trendsHtml = `<li class="summary-item">暂无显著趋势分析</li>`;

  let topicsHtml = (summary.top_topics || []).map(t => `
    <span class="topic-tag clickable-topic" onclick="window.searchByTopic('${t.topic.replace(/'/g, "\\'")}')" title="点击搜索此主题">
        ${t.topic}<span class="count">${t.count}</span>
    </span>
  `).join("");

  let breakthroughsHtml = (summary.breakthroughs || []).map(p => `
    <div class="highlight-card">
      <div class="highlight-title">${p.title}</div>
      <div class="highlight-desc">"${p.one_liner || "摘要详见原文"}"</div>
      <div class="highlight-meta">${formatAuthors(p.authors)} · ${p.venue || "arXiv"}</div>
      <div class="tags" style="margin-top:4px;">
         ${p.ccf_class && p.ccf_class !== 'None' ? `<span class="ccf-tag ccf-${p.ccf_class}">CCF ${p.ccf_class}</span>` : ''}
         <a href="${p.url}" target="_blank" class="read-btn" style="font-size:12px;">阅读原文</a>
      </div>
    </div>
  `).join("");
  
  if (!breakthroughsHtml) breakthroughsHtml = `<div class="highlight-meta">今日暂无高亮突破论文</div>`;

  const html = `
    <div class="daily-header">
      <div class="daily-title">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
          <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
        </svg>
        今日 AI 领域前沿
      </div>
      <div class="daily-date">${dateStr} · 今日新增论文 ${summary.total_papers} 篇</div>
    </div>
    
    <div class="daily-content">
      <!-- Left Column: Trends & Breakthroughs -->
      <div class="summary-column">
        <div class="summary-section">
          <h3>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="20" x2="18" y2="10"></line>
              <line x1="12" y1="20" x2="12" y2="4"></line>
              <line x1="6" y1="20" x2="6" y2="14"></line>
            </svg>
            关键趋势与突破
          </h3>
          <ul class="summary-list">
            ${trendsHtml}
          </ul>
        </div>
        
        <div class="summary-section" style="margin-top: 24px;">
           <h3>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>
            </svg>
             精选亮点论文
           </h3>
           <div class="highlight-list">
             ${breakthroughsHtml}
           </div>
        </div>
      </div>
      
      <!-- Right Column: Hot Topics -->
      <div class="summary-column">
        <div class="summary-section">
          <h3>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20.24 12.24a6 6 0 0 0-8.49-8.49L5 10.5V19h8.5z"></path>
              <line x1="16" y1="8" x2="2" y2="22"></line>
              <line x1="17.5" y1="15" x2="9" y2="15"></line>
            </svg>
            高频研究主题
          </h3>
          <div class="topic-tags">
            ${topicsHtml}
          </div>
        </div>
      </div>
    </div>
  `;
  
  container.innerHTML = html;
}

// Event Listeners
refreshBtn.addEventListener("click", () => fetchPapers(false));
queryInput.addEventListener("keypress", (e) => {
    if (e.key === "Enter") fetchPapers(false);
});

// Filters trigger immediate update (client-side filtering or re-fetch)
// For deep filters (server side), we might want to re-fetch
categorySelect.addEventListener("change", applyFilters);
sortSelect.addEventListener("change", () => {
    activeSort = sortSelect.value;
    fetchPapers(false);
});

// For these, we should probably re-fetch because they affect the dataset from server
monthInput.addEventListener("change", () => fetchPapers(false));
ccfSelect.addEventListener("change", () => fetchPapers(false));
topTierCheckbox.addEventListener("change", () => fetchPapers(false));
sourceCheckboxes.forEach(cb => cb.addEventListener("change", () => fetchPapers(false)));

// Initial load
fetchDailySummary();
// fetchPapers(false); // Optional: don't auto-fetch search results to keep it clean, or do it if user wants to see list below 
