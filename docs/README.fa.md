<div dir="rtl" align="center">
  <h1>🔥 FireScan</h1>
  <h3>ابزار پیشرفته اسکن پورت و شناسایی شبکه</h3>
  <p>
    <a href="https://github.com/mr-coder20/FireScan/releases"><img src="https://img.shields.io/github/v/release/mr-coder20/FireScan?style=flat-square&color=orange" alt="Release"></a>
    <a href="https://github.com/mr-coder20/FireScan/actions"><img src="https://img.shields.io/github/actions/workflow/status/mr-coder20/FireScan/test.yml?style=flat-square" alt="CI"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/mr-coder20/FireScan?style=flat-square" alt="License"></a>
  </p>
  <p><strong>"سرعت Masscan ❄ عمق Nmap"</strong></p>
</div>

<div dir="rtl">

---

## ✨ قابلیت‌ها

| ویژگی | توضیحات |
|---|---|
| 🚀 **موتور ترکیبی** | انتخاب هوشمند بین Masscan، RustScan، Nmap و موتور خالص Go |
| ⚡ **سرعت Masscan** | اسکن میلیون‌ها IP در چند دقیقه |
| 🎯 **سرعت RustScan** | اسکن ۶۵٬۰۰۰ پورت در ~۳ ثانیه |
| 🔬 **عمق Nmap** | ۲۲٬۰۰۰+ اثر انگشت سرویس، اسکریپت‌های NSE، تشخیص OS |
| 📊 **خروجی‌های متنوع** | جدول، JSON، CSV، HTML، Greppable |
| 🔧 **حالت Pipe** | کار در اسکریپت‌ها و پایپ‌لاین‌های CI/CD |
| 🌐 **چندسکویی** | لینوکس، macOS، ویندوز، داکر |
| 🧩 **بدون وابستگی** | موتور خالص Go بدون نیاز به ابزار خارجی |

---

## 📦 نصب

### نصب یک‌خطی (لینوکس/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/mr-coder20/FireScan/main/scripts/install.sh | bash