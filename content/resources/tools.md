# اکوسیستم ابزارها و ارائه‌دهندگان AI
## راهنمای انتخاب برای کارگاه‌ها، نه فهرست وابسته به برند

ابتدا قابلیت و مسئله را مشخص می‌کنیم، سپس ابزار را انتخاب می‌کنیم. «متن‌باز»، «وزن‌باز»، «رایگان» و «دارای پلن رایگان» یک معنی ندارند:
- **Open source:** کد با مجوز مشخص قابل بررسی، استفاده و تغییر است.
- **Open weights:** وزن مدل قابل دریافت است؛ به‌تنهایی به معنی متن‌باز بودن داده، آموزش یا مجوز آزاد نیست.
- **Free to use:** ممکن است رایگان اما اختصاصی و محدود باشد.
- **Freemium / free tier:** استفاده رایگان با سقف، محدودیت مدل یا شرایط متغیر.
- **Trial credit:** اعتبار آزمایشی یک‌باره، نه پلن رایگان پایدار.

## ۱. مدل‌ها و ارائه‌دهندگان API

این گزینه‌ها برای نمونه‌سازی، مقایسه مدل و اتصال API قابل بررسی‌اند. سهمیه‌ها و مدل‌های رایگان سریع تغییر می‌کنند؛ بنابراین این راهنما عمداً جدول ثابت سهمیه‌ها یا دسترسی کشورها ندارد. پیش از هر اجرا، مستندات رسمی و داشبورد حساب را بررسی کنید.

- **[Google AI Studio / Gemini API](https://ai.google.dev/gemini-api/docs):** محیط آزمایش مدل‌های Gemini و API با tier رایگان برای مدل‌های منتخب. سقف‌ها به مدل و پروژه وابسته‌اند؛ داده‌های tier رایگان ممکن است برای بهبود محصولات Google استفاده شوند. داده محرمانه وارد نکنید.
- **[GroqCloud](https://console.groq.com/docs/overview):** inference سریع برای مدل‌های open-weight با API سازگار با الگوهای رایج OpenAI. پلن رایگان مدل‌محور است؛ سقف واقعی در صفحه Limits حساب دیده می‌شود.
- **[Cerebras Inference](https://inference-docs.cerebras.ai/):** گزینه‌ای برای تجربه inference سریع. اعتبار آزمایشی و دسترسی مدل‌ها را با free tier دائمی اشتباه نگیرید.
- **[Cloudflare Workers AI](https://developers.cloudflare.com/workers-ai/):** اجرای مدل‌های منتخب از طریق API و اکوسیستم Workers؛ سهمیه رایگان روزانه بر حسب Neurons تعریف شده و برخی مدل‌ها به پلن یا پرداخت نیاز دارند.
- **[OpenRouter](https://openrouter.ai/docs):** درگاه واحد برای مقایسه و استفاده از مدل‌های متعدد. مدل‌های دارای شناسه :free سهمیه جداگانه دارند؛ provider پشت مدل، سیاست داده و ظرفیت آن اهمیت دارد.
- **[SambaNova Cloud](https://docs.sambanova.ai/):** API برای مدل‌های منتخب و مسیر آزمایش رایگان با محدودیت روزانه و مدل‌محور؛ headerهای نرخ محدودیت را بررسی کنید.
- **[NVIDIA API Catalog](https://build.nvidia.com/explore/discover):** محیط آزمایش endpointهای مدل و نمونه‌های NIM؛ شرایط هر endpoint و مجوز استفاده تولیدی را جداگانه بخوانید.
- **[Cohere](https://docs.cohere.com/docs/rate-limits):** کلیدهای ارزیابی رایگان با محدودیت تعداد درخواست برای آزمایش RAG، embedding و reranking؛ کلید trial برای production یا استفاده تجاری طراحی نشده است.
- **[Mistral AI](https://docs.mistral.ai/):** گزینه مهم برای مدل‌ها و ابزارهای توسعه؛ طبق مستندات فعلی، Free mode برای تجربه Vibe و Studio است و شامل API نمی‌شود؛ API نیازمند فعال‌سازی billing است.
- **[Hugging Face](https://huggingface.co/docs/inference-providers/index):** نقطه اصلی کشف مدل‌ها، دیتاست‌ها و providerهای گوناگون. دسترسی Inference Providers را رایگان فرض نکنید؛ اعتبار و شرایط به حساب و سرویس متصل وابسته است.
- **[Ollama Cloud](https://ollama.com/):** گزینه ابری برای تجربه مدل‌ها؛ پلن و سهمیه آن را جدا بررسی کنید. با اجرای محلی Ollama یکی نیست.

هر راهنمای قدیمی درباره GitHub Models، endpointها یا سهمیه رایگان آن باید با وضعیت رسمی روز تطبیق داده شود. از snippetهای قدیمی به‌عنوان زیرساخت کلاس استفاده نکنید.

### پروتکل بررسی provider پیش از هر نوبت
1. صفحه رسمی pricing، rate limits و terms را باز کنید.
2. مشخص کنید دسترسی رایگان دائمی است یا اعتبار آزمایشی.
3. نام دقیق مدل و تاریخ بررسی را ثبت کنید.
4. سقف RPM/TPM/RPD، محدودیت هم‌زمانی و نیاز به کارت یا احراز هویت را بررسی کنید.
5. سیاست نگهداری داده، استفاده برای آموزش و محدودیت استفاده تجاری را بخوانید.
6. یک درخواست آزمایشی بی‌خطر اجرا کنید و خطاهای 429، timeout و پاسخ ناقص را بررسی کنید.
7. provider دوم یا مدل جایگزین آماده کنید.

### انتخاب مدل به‌جای انتخاب برند
کیفیت استدلال، فارسی، کدنویسی، ورودی چندوجهی، context، latency، پایداری، هزینه، مجوز و سیاست داده را متناسب با مسئله بسنجید. تمرین پیشنهادی: سه مدل را با prompt و rubric یکسان آزمایش کنید و خروجی‌ها را با معیار مقایسه کنید، نه حس کلی «این یکی باهوش‌تر است».

## ۲. اجرای محلی و مدل‌های open-weight
- **[Ollama](https://github.com/ollama/ollama):** شروع ساده برای اجرای مدل محلی و ارائه API.
- **[llama.cpp](https://github.com/ggml-org/llama.cpp):** اجرای مدل‌های GGUF با کنترل بیشتر روی CPU/GPU و quantization.
- **[LM Studio](https://lmstudio.ai/):** رابط دسکتاپ برای دریافت و آزمایش مدل‌های محلی؛ رایگان اما اختصاصی.
- **[Jan](https://github.com/JanAI-ai/jan):** دستیار دسکتاپ local-first.
- **[LocalAI](https://github.com/mudler/LocalAI):** API سازگار برای اجرای محلی مدل‌های متنوع.
- **[vLLM](https://github.com/vllm-project/vllm):** serving با throughput بالا؛ مناسب مسیر سرور و پیشرفته.
- **[SGLang](https://github.com/sgl-project/sglang):** runtime و serving پیشرفته.
- **[MLC LLM](https://github.com/mlc-ai/mlc-llm):** کامپایل و اجرای مدل روی پلتفرم‌های مختلف.
- **[Open WebUI](https://github.com/open-webui/open-webui):** رابط self-host برای مدل‌های محلی و سرویس‌های سازگار.
- **[llamafile](https://github.com/Mozilla-Ocho/llamafile):** بسته‌بندی ساده‌تر مدل و runtime.

اجرای محلی به معنی رایگان بودن محاسبات، برق، حافظه یا فضای ذخیره‌سازی نیست. برای جلسه مقدماتی نمایش مدل کوچک یا از پیش آماده کافی است.

## ۳. آفرینش دیجیتال: تصویر، طراحی، ویدیو، صدا

### تصویر و طراحی
- **[ComfyUI](https://github.com/comfyanonymous/ComfyUI):** محیط node-based برای pipeline تصویر، ویدیو و صوت؛ مناسب نشان دادن فرایند تولید قابل مشاهده.
- **[InvokeAI](https://github.com/invoke-ai/InvokeAI):** canvas، inpainting و workflow برای تولید و اصلاح تصویر.
- **[Krita](https://krita.org/) + [Krita AI Diffusion](https://github.com/Acly/krita-ai-diffusion):** پیوند نقاشی دیجیتال با pipelineهای تولید و ویرایش تصویر.
- **[GIMP](https://www.gimp.org/):** ویرایش تصویر متن‌باز و محلی.
- **[Penpot](https://penpot.app/):** طراحی رابط و prototype.
- **[Inkscape](https://inkscape.org/):** طراحی برداری.
- **[Blender](https://www.blender.org/):** مدل‌سازی سه‌بعدی، انیمیشن و رندر.
- **[Photopea](https://www.photopea.com/):** ویرایشگر مرورگری برای PSD و تصویر؛ رایگان برای استفاده پایه اما متن‌باز نیست.

### ویدیو و تصویر متحرک
- **[Kdenlive](https://kdenlive.org/)** و **[Shotcut](https://shotcut.org/):** تدوین ویدیو.
- **[OBS Studio](https://obsproject.com/):** ضبط صفحه و تولید محتوای آموزشی.
- **[FFmpeg](https://ffmpeg.org/):** پردازش، تبدیل و ترکیب رسانه.
- **[OpenTimelineIO](https://github.com/AcademySoftwareFoundation/OpenTimelineIO):** تبادل ساختار timeline.

### صوت، گفتار و موسیقی
- **[Audacity](https://www.audacityteam.org/):** ضبط و ویرایش صوت.
- **[Whisper](https://github.com/openai/whisper):** تشخیص گفتار و تبدیل صوت به متن.
- **[WhisperX](https://github.com/m-bain/whisperX):** هم‌ترازی و timestamp برای transcription.
- **[Piper](https://github.com/rhasspy/piper):** تبدیل متن به گفتار محلی و سبک.
- **[Kokoro](https://github.com/hexgrad/kokoro):** مدل TTS با وزن‌های قابل دریافت؛ مجوز کد و وزن‌ها را جدا بررسی کنید.
- **[Demucs](https://github.com/facebookresearch/demucs):** جداسازی stemهای صوتی.
- **[RVC](https://github.com/RVC-Project/Retrieval-based-Voice-Conversion-WebUI):** تبدیل صدا؛ استفاده از صدای افراد واقعی نیازمند رضایت و رعایت حقوق آن‌هاست.
- **[AudioCraft](https://github.com/facebookresearch/audiocraft):** ابزارهای پژوهشی تولید صوت و موسیقی؛ مجوز کد و وزن‌ها جداگانه بررسی شود.

تمرین Creative: یک داستان ۳۰ ثانیه‌ای را به storyboard، تصویر، صدای روایت و ویدیوی کوتاه تبدیل کنید. انسجام روایت و تصمیم هنری معیار اصلی است.

## ۴. توسعه، برنامه‌نویسی و نمونه‌سازی
- **[Continue](https://github.com/continuedev/continue):** دستیار کدنویسی قابل تنظیم.
- **[Aider](https://github.com/Aider-AI/aider):** برنامه‌نویسی جفتی در ترمینال.
- **[OpenHands](https://github.com/All-Hands-AI/OpenHands):** agent برای وظایف توسعه نرم‌افزار.
- **[OpenCode](https://github.com/anomalyco/opencode):** عامل کدنویسی در ترمینال.
- **[Pydantic AI](https://github.com/pydantic/pydantic-ai):** ساخت agentهای Python با type safety و providerهای مختلف.
- **[DSPy](https://github.com/stanfordnlp/dspy):** طراحی و بهینه‌سازی pipelineهای LLM.
- **[FastAPI](https://github.com/fastapi/fastapi):** ساخت API در Python.
- **[Vite](https://github.com/vitejs/vite):** شروع سریع frontend.
- **[Astro](https://github.com/withastro/astro):** وب‌سایت‌های محتوایی سریع.
- **[Git](https://git-scm.com/):** نسخه‌بندی و ثبت تغییرات.

## ۵. Workflow، Agent و ابزارهای بصری
- **[n8n](https://github.com/n8n-io/n8n):** اتصال سرویس‌ها و workflow خودکار؛ مجوز و مدل استقرار را بخوانید.
- **[Dify](https://github.com/langgenius/dify):** ساخت برنامه‌های LLM، RAG و workflow با رابط بصری.
- **[Flowise](https://github.com/FlowiseAI/Flowise):** طراحی بصری pipeline و agent.
- **[Langflow](https://github.com/langflow-ai/langflow):** طراحی flowهای AI.
- **[LangGraph](https://github.com/langchain-ai/langgraph):** orchestration stateful برای agentها.
- **[LlamaIndex](https://github.com/run-llama/llama_index):** اتصال مدل به داده و RAG.
- **[Haystack](https://github.com/deepset-ai/haystack):** pipeline جستجو و RAG.
- **[AG2](https://github.com/ag2ai/ag2):** orchestration و همکاری agentها.
- **[MCP](https://modelcontextprotocol.io/):** پروتکل اتصال مدل‌ها به ابزارها و منابع خارجی؛ امنیت و مجوز ابزارها باید آموزش داده شود.

## ۶. RAG، جستجو و دانش
- **[Qdrant](https://github.com/qdrant/qdrant):** پایگاه برداری و جستجوی ترکیبی.
- **[Chroma](https://github.com/chroma-core/chroma):** پایگاه embedding برای نمونه‌سازی.
- **[FAISS](https://github.com/facebookresearch/faiss):** جستجوی برداری.
- **[Sentence Transformers](https://github.com/UKPLab/sentence-transformers):** embedding و semantic similarity.
- **[Unstructured](https://github.com/Unstructured-IO/unstructured):** استخراج و آماده‌سازی داده از اسناد.
- **[Docling](https://github.com/docling-project/docling):** پردازش و تبدیل اسناد.
- **[OpenAlex](https://openalex.org/)**، **[Crossref](https://www.crossref.org/)** و **[Semantic Scholar](https://www.semanticscholar.org/):** کشف منابع علمی و فراداده پژوهش.

## ۷. ارزیابی، مشاهده‌پذیری و امنیت
- **[Promptfoo](https://github.com/promptfoo/promptfoo):** تست prompt، مقایسه مدل و red-teaming.
- **[DeepEval](https://github.com/confident-ai/deepeval):** ارزیابی خروجی برنامه‌های LLM.
- **[Ragas](https://github.com/explodinggradients/ragas):** ارزیابی RAG.
- **[Langfuse](https://github.com/langfuse/langfuse):** tracing و مدیریت prompt.
- **[Arize Phoenix](https://github.com/Arize-ai/phoenix):** tracing و ارزیابی.
- **[OpenTelemetry](https://opentelemetry.io/):** استاندارد telemetry.
- **[OWASP Top 10 for LLM Applications](https://owasp.org/www-project-top-10-for-large-language-model-applications/):** چارچوب تهدیدهای برنامه‌های LLM.

## ۸. داده، پژوهش و یادگیری
- **[JupyterLab](https://github.com/jupyterlab/jupyterlab):** محیط محاسبات تعاملی.
- **[pandas](https://github.com/pandas-dev/pandas)** و **[Polars](https://github.com/pola-rs/polars):** پردازش داده.
- **[scikit-learn](https://github.com/scikit-learn/scikit-learn):** یادگیری ماشین کلاسیک.
- **[Hugging Face Transformers](https://github.com/huggingface/transformers):** مدل‌ها و pipelineهای یادگیری عمیق.
- **[Hugging Face Diffusers](https://github.com/huggingface/diffusers):** مدل‌ها و pipelineهای diffusion.
- **[Microsoft AI for Beginners](https://github.com/microsoft/AI-For-Beginners):** مسیر مقدماتی با درس و تمرین.
- **[Hugging Face Learn](https://huggingface.co/learn):** دوره‌ها و منابع فنی.

## ۹. تبدیل ابزار به تجربه آموزشی

هر ابزار فقط وقتی وارد برنامه می‌شود که نقش آموزشی روشن داشته باشد. برای هر ابزار ثبت کنید:
- مسئله‌ای که حل می‌کند.
- مفهوم فنی یا خلاقانه‌ای که آموزش می‌دهد.
- خروجی قابل ساخت در ۳۰ تا ۹۰ دقیقه.
- پیش‌نیاز دانشجو.
- مجوز کد، وزن مدل، داده و دارایی تولیدی.
- هزینه، سیاست داده و شرایط استفاده.
- جایگزینی که مفهوم آموزشی را ثابت نگه می‌دارد.

در یک جلسه بیش از یک ابزار اصلی و یک ابزار مکمل معرفی نکنید، مگر اینکه خود مقایسه موضوع یادگیری باشد. به‌جای آموزش سطحی ۱۵ ابزار، یک خروجی واقعی را با یک ابزار بسازید و سپس جایگزین را نشان دهید.

## منابع رسمی providerها برای بررسی روز اجرا
- [Gemini API rate limits](https://ai.google.dev/gemini-api/docs/rate-limits)
- [Gemini API pricing](https://ai.google.dev/gemini-api/docs/pricing)
- [Groq rate limits](https://console.groq.com/docs/rate-limits)
- [OpenRouter API limits](https://openrouter.ai/docs/api_reference/limits)
- [Cloudflare Workers AI pricing](https://developers.cloudflare.com/workers-ai/platform/pricing/)
- [SambaNova rate limits](https://docs.sambanova.ai/docs/en/models/rate-limits)
- [Cohere API rate limits](https://docs.cohere.com/docs/rate-limits)
- [Mistral usage and limits](https://docs.mistral.ai/admin/billing-usage/usage-limits)
- [NVIDIA API Catalog](https://build.nvidia.com/explore/discover)

**آخرین بازبینی مستندات providerها: ۱۲ نوامبر ۲۰۲۶.** این تاریخ تضمین دسترسی یا سهمیه برای همه حساب‌ها نیست؛ داشبورد حساب و مستندات رسمی جاری مرجع نهایی‌اند.
